package resource

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
)

type (
	// CronJob tracks a kubernetes resource.
	CronJob struct {
		*Base
		instance *batchv1beta1.CronJob
	}

	// Runner can run jobs.
	Runner interface {
		Run(path string) error
	}

	// Runnable can run jobs.
	Runnable interface {
		Run(ns, n string) error
	}
)

// NewCronJobList returns a new resource list.
func NewCronJobList(ns string) List {
	return NewCronJobListWithArgs(ns, NewCronJob())
}

// NewCronJobListWithArgs returns a new resource list.
func NewCronJobListWithArgs(ns string, res Resource) List {
	return newList(ns, "cronjob", res, AllVerbsAccess|DescribeAccess)
}

// NewCronJob instantiates a new CronJob.
func NewCronJob() *CronJob {
	return NewCronJobWithArgs(k8s.NewCronJob())
}

// NewCronJobWithArgs instantiates a new CronJob.
func NewCronJobWithArgs(r k8s.Res) *CronJob {
	cm := &CronJob{
		Base: &Base{
			caller: r,
		},
	}
	cm.creator = cm
	return cm
}

// NewInstance builds a new CronJob instance from a k8s resource.
func (*CronJob) NewInstance(i interface{}) Columnar {
	job := NewCronJob()
	switch i.(type) {
	case *batchv1beta1.CronJob:
		job.instance = i.(*batchv1beta1.CronJob)
	case batchv1beta1.CronJob:
		ii := i.(batchv1beta1.CronJob)
		job.instance = &ii
	default:
		log.Fatalf("Unknown %#v", i)
	}
	job.path = job.namespacedName(job.instance.ObjectMeta)
	return job
}

// Marshal resource to yaml.
func (r *CronJob) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
	if err != nil {
		return "", err
	}

	cj := i.(*batchv1beta1.CronJob)
	cj.TypeMeta.APIVersion = "extensions/batchv1beta1"
	cj.TypeMeta.Kind = "CronJob"
	return r.marshalObject(cj)
}

// Run a given cronjob.
func (r *CronJob) Run(pa string) error {
	ns, n := namespaced(pa)
	if c, ok := r.caller.(Runnable); ok {
		return c.Run(ns, n)
	}
	return fmt.Errorf("unable to run cronjob %s", pa)
}

// Header return resource header.
func (*CronJob) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}
	return append(hh, "NAME", "SCHEDULE", "SUSPEND", "ACTIVE", "LAST_SCHEDULE", "AGE")
}

// Fields retrieves displayable fields.
func (r *CronJob) Fields(ns string) Row {
	ff := make([]string, 0, len(r.Header(ns)))

	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	lastScheduled := "<none>"
	if i.Status.LastScheduleTime != nil {
		lastScheduled = toAge(*i.Status.LastScheduleTime)
	}

	return append(ff,
		i.Name,
		i.Spec.Schedule,
		boolToStr(*i.Spec.Suspend),
		strconv.Itoa(len(i.Status.Active)),
		lastScheduled,
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ExtFields returns extended fields in relation to headers.
func (*CronJob) ExtFields() Properties {
	return Properties{}
}
