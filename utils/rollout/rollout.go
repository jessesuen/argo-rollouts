package rollout

// import (
// 	"k8s.io/client-go/rest"
// 	"k8s.io/client-go/dynamic"
// 	"k8s.io/client-go/dynamic/dynamicinformer"

// 	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts"

// )

// // NewRolloutInformer returns a informer used by the controller. This uses a dynamic informer to
// // work around https://github.com/kubernetes/kubernetes/issues/57705
// func NewRolloutInformer(cfg *rest.Config, ns string, resyncPeriod time.Duration) cache.SharedIndexInformer {
// 	dclient, err := dynamic.NewForConfig(cfg)
// 	if err != nil {
// 		panic(err)
// 	}
// 	factory := dynamicinformer.NewDynamicSharedInformerFactory(dclient, resyncPeriod)
// 	genericInformer := factory.ForResource(schema.GroupVersionResource{
// 		Group:    rollouts.Group,
// 		Version:  "v1alpha1",
// 		Resource: rollouts.Plural,
// 	})
// 	return genericInformer.Informer()
// }

// // WorkflowLister implements the List() method of v1alpha.WorkflowLister interface but does so using
// // an Unstructured informer and converting objects to workflows. Ignores objects that failed to convert.
// type RolloutLister interface {
// 	List() ([]*wfv1.Workflow, error)
// }

// type workflowLister struct {
// 	informer cache.SharedIndexInformer
// }

// func (l *workflowLister) List() ([]*wfv1.Workflow, error) {
// 	workflows := make([]*wfv1.Workflow, 0)
// 	for _, m := range l.informer.GetStore().List() {
// 		wf, err := FromUnstructured(m.(*unstructured.Unstructured))
// 		if err != nil {
// 			log.Warnf("Failed to unmarshal workflow %v object: %v", m, err)
// 			continue
// 		}
// 		workflows = append(workflows, wf)
// 	}
// 	return workflows, nil
// }

// // NewWorkflowLister returns a new workflow lister
// func NewWorkflowLister(informer cache.SharedIndexInformer) WorkflowLister {
// 	return &workflowLister{
// 		informer: informer,
// 	}
// }
