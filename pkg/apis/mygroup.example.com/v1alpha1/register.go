package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	GroupName = "mygroup.example.com"
	Version   = "v1alpha1"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{
	Group:   GroupName,
	Version: Version,
}

var (
	// SchemeBuilder collects functions that add things to a scheme.
	SchemeBuilder      = runtime.NewSchemeBuilder(addKnownTypes)
	localSchemeBuilder = &SchemeBuilder

	// AddToScheme adds all types of this group/version to a scheme.
	AddToScheme = localSchemeBuilder.AddToScheme
)

// addKnownTypes adds our types to the API scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		SchemeGroupVersion,
		&MyResource{},
		&MyResourceList{},
	)

	// Register the group version with the scheme (for meta types)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

// Kind returns a GroupKind for this API group's kind.
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource returns a GroupResource for this API group's resource.
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
