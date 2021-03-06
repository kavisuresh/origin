package deployerpod

import (
	"fmt"

	"github.com/golang/glog"

	kapi "k8s.io/kubernetes/pkg/api"
	kerrors "k8s.io/kubernetes/pkg/api/errors"

	deployapi "github.com/openshift/origin/pkg/deploy/api"
	deployutil "github.com/openshift/origin/pkg/deploy/util"
)

// DeployerPodController keeps a deployment's status in sync with the deployer pod
// handling the deployment.
//
// Use the DeployerPodControllerFactory to create this controller.
type DeployerPodController struct {
	// deploymentClient provides access to deployments.
	deploymentClient deploymentClient
	// deployerPodsFor returns all deployer pods for the named deployment.
	deployerPodsFor func(namespace, name string) (*kapi.PodList, error)
	// decodeConfig knows how to decode the deploymentConfig from a deployment's annotations.
	decodeConfig func(deployment *kapi.ReplicationController) (*deployapi.DeploymentConfig, error)
	// deletePod deletes a pod.
	deletePod func(namespace, name string) error
}

// transientError is an error which will be retried indefinitely.
type transientError string

func (e transientError) Error() string { return "transient error handling deployer pod: " + string(e) }

// Handle syncs pod's status with any associated deployment.
func (c *DeployerPodController) Handle(pod *kapi.Pod) error {
	// Find the deployment associated with the deployer pod.
	deploymentName := deployutil.DeploymentNameFor(pod)
	if len(deploymentName) == 0 {
		return nil
	}
	// Reject updates to anything but the main deployer pod
	// TODO: Find a way to filter this on the watch side.
	if pod.Name != deployutil.DeployerPodNameForDeployment(deploymentName) {
		return nil
	}

	deployment, err := c.deploymentClient.getDeployment(pod.Namespace, deploymentName)
	// If the deployment for this pod has disappeared, we should clean up this
	// and any other deployer pods, then bail out.
	if err != nil {
		// Some retrieval error occurred. Retry.
		if !kerrors.IsNotFound(err) {
			return fmt.Errorf("couldn't get deployment %s/%s which owns deployer pod %s/%s", pod.Namespace, deploymentName, pod.Name, pod.Namespace)
		}
		// Find all the deployer pods for the deployment (including this one).
		deployers, err := c.deployerPodsFor(pod.Namespace, deploymentName)
		if err != nil {
			// Retry.
			return fmt.Errorf("couldn't get deployer pods for %s: %v", deployutil.LabelForDeployment(deployment), err)
		}
		// Delete all deployers.
		for _, deployer := range deployers.Items {
			err := c.deletePod(deployer.Namespace, deployer.Name)
			if err != nil {
				if !kerrors.IsNotFound(err) {
					// TODO: Should this fire an event?
					glog.V(2).Infof("Couldn't delete orphaned deployer pod %s/%s: %v", deployer.Namespace, deployer.Name, err)
				}
			} else {
				// TODO: Should this fire an event?
				glog.V(2).Infof("Deleted orphaned deployer pod %s/%s", deployer.Namespace, deployer.Name)
			}
		}
		return nil
	}

	currentStatus := deployutil.DeploymentStatusFor(deployment)
	nextStatus := currentStatus

	switch pod.Status.Phase {
	case kapi.PodRunning:
		if !deployutil.IsTerminatedDeployment(deployment) {
			nextStatus = deployapi.DeploymentStatusRunning
		}
	case kapi.PodSucceeded:
		// Detect failure based on the container state
		nextStatus = deployapi.DeploymentStatusComplete
		for _, info := range pod.Status.ContainerStatuses {
			if info.State.Terminated != nil && info.State.Terminated.ExitCode != 0 {
				nextStatus = deployapi.DeploymentStatusFailed
				break
			}
		}
		// Sync the internal replica annotation with the target so that we can
		// distinguish deployer updates from other scaling events.
		deployment.Annotations[deployapi.DeploymentReplicasAnnotation] = deployment.Annotations[deployapi.DesiredReplicasAnnotation]
		if nextStatus == deployapi.DeploymentStatusComplete {
			delete(deployment.Annotations, deployapi.DesiredReplicasAnnotation)
		}

		// reset the size of any test container, since we are the ones updating the RC
		if config, err := c.decodeConfig(deployment); err == nil && config.Spec.Test {
			deployment.Spec.Replicas = 0
		}
	case kapi.PodFailed:
		nextStatus = deployapi.DeploymentStatusFailed

		// reset the size of any test container, since we are the ones updating the RC
		if config, err := c.decodeConfig(deployment); err == nil && config.Spec.Test {
			deployment.Spec.Replicas = 0
		}
	}

	if currentStatus != nextStatus {
		deployment.Annotations[deployapi.DeploymentStatusAnnotation] = string(nextStatus)
		if _, err := c.deploymentClient.updateDeployment(deployment.Namespace, deployment); err != nil {
			if kerrors.IsNotFound(err) {
				return nil
			}
			return fmt.Errorf("couldn't update Deployment %s to status %s: %v", deployutil.LabelForDeployment(deployment), nextStatus, err)
		}
		glog.V(4).Infof("Updated deployment %s status from %s to %s (scale: %d)", deployutil.LabelForDeployment(deployment), currentStatus, nextStatus, deployment.Spec.Replicas)
	}

	return nil
}

// deploymentClient abstracts access to deployments.
type deploymentClient interface {
	getDeployment(namespace, name string) (*kapi.ReplicationController, error)
	updateDeployment(namespace string, deployment *kapi.ReplicationController) (*kapi.ReplicationController, error)
	// listDeploymentsForConfig should return deployments associated with the
	// provided config.
	listDeploymentsForConfig(namespace, configName string) (*kapi.ReplicationControllerList, error)
}

// deploymentClientImpl is a pluggable deploymentControllerDeploymentClient.
type deploymentClientImpl struct {
	getDeploymentFunc            func(namespace, name string) (*kapi.ReplicationController, error)
	updateDeploymentFunc         func(namespace string, deployment *kapi.ReplicationController) (*kapi.ReplicationController, error)
	listDeploymentsForConfigFunc func(namespace, configName string) (*kapi.ReplicationControllerList, error)
}

func (i *deploymentClientImpl) getDeployment(namespace, name string) (*kapi.ReplicationController, error) {
	return i.getDeploymentFunc(namespace, name)
}

func (i *deploymentClientImpl) updateDeployment(namespace string, deployment *kapi.ReplicationController) (*kapi.ReplicationController, error) {
	return i.updateDeploymentFunc(namespace, deployment)
}

func (i *deploymentClientImpl) listDeploymentsForConfig(namespace, configName string) (*kapi.ReplicationControllerList, error) {
	return i.listDeploymentsForConfigFunc(namespace, configName)
}
