package test

import (
	"context"
	"encoding/json"
	"github.com/oam-dev/kubevela-core-api/apis/core.oam.dev/common"
	oamv1beta1 "github.com/oam-dev/kubevela-core-api/apis/core.oam.dev/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic/fake"
	"log"
	"testing"
)

/**
 * @Author kingram
 * @Description use fakeClient to operate kubevela resources
 * @Date 2022/8/10
 **/
func TestFakeDynamicClient_Application(t *testing.T) {
	s := runtime.NewScheme()
	err := oamv1beta1.AddToScheme(s)
	assert.Nil(t, err)
	client := fake.NewSimpleDynamicClient(s)

	var name = "my-application"
	var ns = "ns-1"
	var rs int32= 2
	var components []common.ApplicationComponent
	var traits []common.ApplicationTrait
	rsByte,_ := json.Marshal(rs)
	trait := common.ApplicationTrait{
		Type:	"scaler",
		Properties: &runtime.RawExtension{
			Raw: rsByte,
		},
	}
	traits = append(traits,trait)
	components = append(components,common.ApplicationComponent{
		Traits: traits,
	})

	app := &oamv1beta1.Application{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec : oamv1beta1.ApplicationSpec{
			Components: components,
		},
	}

	gvr := schema.GroupVersionResource{
		Group:    "core.oam.dev",
		Version:  "v1beta1",
		Resource: "applications",
	}

	var appUnstructred = &unstructured.Unstructured{}
	err = s.Convert(app, appUnstructred, context.Background())
	assert.Nil(t, err)

	// create application
	appCreate, err := client.Resource(gvr).Namespace(ns).Create(context.Background(),appUnstructred,metav1.CreateOptions{})
	assert.Nil(t, err)

	log.Printf("appCreate:%+v \n",appCreate.Object)

	var newRs int32 = 1
	newRsByte,_ := json.Marshal(newRs)
	app.Spec.Components[0].Traits[0].Properties =&runtime.RawExtension{
		Raw: newRsByte,
	}
	var newAppUnstructred = &unstructured.Unstructured{}
	err = s.Convert(app, newAppUnstructred, context.Background())
	assert.Nil(t, err)

	// update application
	appUpdate, err := client.Resource(gvr).Namespace(ns).Update(context.Background(),newAppUnstructred,metav1.UpdateOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, appUpdate)

	// get application
	appGet, err := client.Resource(gvr).Namespace(ns).Get(context.Background(),name,metav1.GetOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, appGet)

	log.Printf("appGet:%+v \n",appGet.Object)

	assert.Equal(t, appCreate.GetName(),appGet.GetName())

	var appGetR = &oamv1beta1.Application{}
	err = s.Convert(appGet, appGetR, context.Background())
	assert.Nil(t, err)

	var iirs int32
	_ = json.Unmarshal(appGetR.Spec.Components[0].Traits[0].Properties.Raw,&iirs)
	assert.Equal(t, iirs,newRs)

	// delete application
	err =  client.Resource(gvr).Namespace(ns).Delete(context.Background(), name, metav1.DeleteOptions{})
	assert.Nil(t, err)
}

/**
 * @Author kingram
 * @Description watch event of kubevela resources
 * @Date 2022/8/10
 **/
func TestFakeDynamicClient_WatchApplication(t *testing.T) {
	s := runtime.NewScheme()
	err := oamv1beta1.AddToScheme(s)
	assert.Nil(t, err)
	client := fake.NewSimpleDynamicClient(s)

	stopChain := make(chan struct{})

	var name = "my-application"
	var ns = "ns-1"
	var rs int32= 2
	var components []common.ApplicationComponent
	var traits []common.ApplicationTrait
	rsByte,_ := json.Marshal(rs)
	trait := common.ApplicationTrait{
		Type:	"scaler",
		Properties: &runtime.RawExtension{
			Raw: rsByte,
		},
	}
	traits = append(traits,trait)
	components = append(components,common.ApplicationComponent{
		Traits: traits,
	})

	app := &oamv1beta1.Application{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app":"1",
			},
		},
		Spec : oamv1beta1.ApplicationSpec{
			Components: components,
		},
	}

	gvr := schema.GroupVersionResource{
		Group:    "core.oam.dev",
		Version:  "v1beta1",
		Resource: "applications",
	}

	appWatch,err := client.Resource(gvr).Namespace(ns).Watch(context.Background(),metav1.ListOptions{
		LabelSelector: "app=1",
	})
	assert.Nil(t, err)
	assert.NotNil(t, appWatch)

	go func() {
		for {
			select {
			case e,ok:= <- appWatch.ResultChan():
				if ok {
					assert.Equal(t,e.Type,watch.Added)
					log.Printf("Got Event:%+v \n",e)
					stopChain <- struct{}{}
				}
			}
		}
	}()

	// create application
	var appUnstructred = &unstructured.Unstructured{}
	err = s.Convert(app, appUnstructred, context.Background())
	assert.Nil(t, err)
	dpCreate,err := client.Resource(gvr).Namespace(ns).Create(context.Background(),appUnstructred,metav1.CreateOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, dpCreate)

	<- stopChain
}