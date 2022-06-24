package log

import (
	"flag"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	v1alpha1 "monitoring/api/v1"
)

const (
	// MetricRunKey defines the key for the metricrun field
	MetricRunKey = "metricrun"
	// NamespaceKey defines the key for the namespace field
	NamespaceKey = "namespace"
)

// SetKLogLevel set the klog level for the k8s go-client
func SetKLogLevel(klogLevel int) {
	klog.InitFlags(nil)
	_ = flag.Set("logtostderr", "true")
	_ = flag.Set("v", strconv.Itoa(klogLevel))
}

// WithObject returns a logging context for an object which includes <kind>=<name> and namespace=<namespace>
func WithObject(obj runtime.Object) *log.Entry {
	logCtx := log.NewEntry(log.StandardLogger())
	gvk := obj.GetObjectKind().GroupVersionKind()
	kind := gvk.Kind
	if kind == "" {
		// it's possible for kind can be empty
		switch obj.(type) {
		case *v1alpha1.MetricRun:
			kind = "metricrun"
		}
	}
	objectMeta, err := meta.Accessor(obj)
	if err == nil {
		logCtx = logCtx.WithField("namespace", objectMeta.GetNamespace())
		logCtx = logCtx.WithField(strings.ToLower(kind), objectMeta.GetName())
	}
	return logCtx
}

// KindNamespaceName is a helper to get kind, namespace, name from a logging context
// This is an optimization that callers can use to avoid inferring this again from a runtime.Object
func KindNamespaceName(logCtx *log.Entry) (string, string, string) {
	var kind string
	var nameIf interface{}
	var ok bool
	if nameIf, ok = logCtx.Data["metricrun"]; ok {
		kind = "MetricRun"
	}
	name, _ := nameIf.(string)
	namespace, _ := logCtx.Data["namespace"].(string)
	return kind, namespace, name
}

// WithMetricRun returns a logging context for MetricRun
func WithMetricRun(ar *v1alpha1.MetricRun) *log.Entry {
	return log.WithField(MetricRunKey, ar.Name).WithField(NamespaceKey, ar.Namespace)
}

// WithRedactor returns a log entry with the inputted secret values redacted
func WithRedactor(entry log.Entry, secrets []string) *log.Entry {
	newFormatter := RedactorFormatter{
		entry.Logger.Formatter,
		secrets,
	}
	entry.Logger.SetFormatter(&newFormatter)
	return &entry
}
