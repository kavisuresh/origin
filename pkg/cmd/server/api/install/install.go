package install

import (
	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/unversioned"

	configapi "github.com/openshift/origin/pkg/cmd/server/api"
	configapiv1 "github.com/openshift/origin/pkg/cmd/server/api/v1"

	_ "github.com/openshift/origin/pkg/build/admission/defaults/api/install"
	_ "github.com/openshift/origin/pkg/build/admission/overrides/api/install"
	_ "github.com/openshift/origin/pkg/project/admission/requestlimit/api/install"
	_ "github.com/openshift/origin/pkg/quota/admission/clusterresourceoverride/api/install"
	_ "github.com/openshift/origin/pkg/quota/admission/runonceduration/api/install"
)

const importPrefix = "github.com/openshift/origin/pkg/cmd/server/api"

var accessor = meta.NewAccessor()

// availableVersions lists all known external versions for this group from most preferred to least preferred
var availableVersions = []unversioned.GroupVersion{configapiv1.SchemeGroupVersion}

func init() {
	if err := enableVersions(availableVersions); err != nil {
		panic(err)
	}
}

// TODO: enableVersions should be centralized rather than spread in each API
// group.
// We can combine registered.RegisterVersions, registered.EnableVersions and
// registered.RegisterGroup once we have moved enableVersions there.
func enableVersions(externalVersions []unversioned.GroupVersion) error {
	addVersionsToScheme(externalVersions...)
	return nil
}

func addVersionsToScheme(externalVersions ...unversioned.GroupVersion) {
	// add the internal version to Scheme
	configapi.AddToScheme(configapi.Scheme)
	// add the enabled external versions to Scheme
	for _, v := range externalVersions {
		switch v {
		case configapiv1.SchemeGroupVersion:
			configapiv1.AddToScheme(configapi.Scheme)

		default:
			glog.Errorf("Version %s is not known, so it will not be added to the Scheme.", v)
			continue
		}
	}
}
