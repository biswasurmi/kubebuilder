package v1

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"k8s.io/apimachinery/pkg/runtime"
)

// Ensure Guestbook implements the webhook interfaces
var _ admission.CustomDefaulter = &Guestbook{}
var _ admission.CustomValidator = &Guestbook{}

// SetupWebhookWithManager registers the webhook with the manager
func (r *Guestbook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(r). // mutating
		WithValidator(r). // validating
		Complete()
}

/////////////////////////
// Mutating Webhook
/////////////////////////
func (r *Guestbook) Default(ctx context.Context, obj runtime.Object) error {
	guestbook, ok := obj.(*Guestbook)
	if !ok {
		return fmt.Errorf("expected a Guestbook but got %T", obj)
	}

	if guestbook.Spec.Size < 1 {
		guestbook.Spec.Size = 1
	}
	if guestbook.Spec.Port == 0 {
		guestbook.Spec.Port = 8080
	}
	if guestbook.Spec.Auth == nil {
		defaultAuth := false
		guestbook.Spec.Auth = &defaultAuth
	}

	return nil
}

/////////////////////////
// Validating Webhook
/////////////////////////
func (r *Guestbook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	guestbook, ok := obj.(*Guestbook)
	if !ok {
		return nil, fmt.Errorf("expected a Guestbook but got %T", obj)
	}

	if guestbook.Spec.Size < 1 {
		return nil, fmt.Errorf("spec.size must be >= 1")
	}
	if guestbook.Spec.Image == "" {
		return nil, fmt.Errorf("spec.image must be provided")
	}

	return nil, nil
}

func (r *Guestbook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	guestbook, ok := newObj.(*Guestbook)
	if !ok {
		return nil, fmt.Errorf("expected a Guestbook but got %T", newObj)
	}

	if guestbook.Spec.Image == "" {
		return nil, fmt.Errorf("spec.image cannot be empty")
	}

	return nil, nil
}

func (r *Guestbook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}
