/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	webappv1 "my.domain/guestbook/api/v1"
)

var guestbooklog = logf.Log.WithName("guestbook-resource")

func SetupGuestbookWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&webappv1.Guestbook{}).
		WithValidator(&GuestbookCustomValidator{}).
		WithDefaulter(&GuestbookCustomDefaulter{}).
		Complete()
}

type GuestbookCustomDefaulter struct {
}

var _ webhook.CustomDefaulter = &GuestbookCustomDefaulter{}

func (d *GuestbookCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	guestbook, ok := obj.(*webappv1.Guestbook)
	if !ok {
		return fmt.Errorf("expected a Guestbook object but got %T", obj)
	}
	guestbooklog.Info("Defaulting for Guestbook", "name", guestbook.GetName())

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

type GuestbookCustomValidator struct {
}

var _ webhook.CustomValidator = &GuestbookCustomValidator{}

func (v *GuestbookCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	guestbook, ok := obj.(*webappv1.Guestbook)
	if !ok {
		return nil, fmt.Errorf("expected a Guestbook object but got %T", obj)
	}
	guestbooklog.Info("Validation for Guestbook upon creation", "name", guestbook.GetName())

	if guestbook.Spec.Size < 1 {
		return nil, fmt.Errorf("spec.size must be >= 1")
	}
	if guestbook.Spec.Image == "" {
		return nil, fmt.Errorf("spec.image must be provided")
	}

	return nil, nil
}

func (v *GuestbookCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	guestbook, ok := newObj.(*webappv1.Guestbook)
	if !ok {
		return nil, fmt.Errorf("expected a Guestbook object for the newObj but got %T", newObj)
	}
	guestbooklog.Info("Validation for Guestbook upon update", "name", guestbook.GetName())

	if guestbook.Spec.Image == "" {
		return nil, fmt.Errorf("spec.image cannot be empty")
	}

	return nil, nil
}

func (v *GuestbookCustomValidator) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	guestbook, ok := obj.(*webappv1.Guestbook)
	if !ok {
		return nil, fmt.Errorf("expected a Guestbook object but got %T", obj)
	}
	guestbooklog.Info("Validation for Guestbook upon deletion", "name", guestbook.GetName())

	return nil, nil
}