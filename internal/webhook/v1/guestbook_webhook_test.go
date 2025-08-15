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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	webappv1 "my.domain/guestbook/api/v1"
)

var _ = Describe("Guestbook Webhook", func() {
	var (
		obj       *webappv1.Guestbook
		oldObj    *webappv1.Guestbook
		ctx       context.Context
		defaulter *GuestbookCustomDefaulter
		validator *GuestbookCustomValidator
	)

	BeforeEach(func() {
		ctx = context.Background()
		obj = &webappv1.Guestbook{}
		oldObj = &webappv1.Guestbook{}
		defaulter = &GuestbookCustomDefaulter{}
		validator = &GuestbookCustomValidator{}
	})

	Context("Defaulting Webhook", func() {
		It("should apply default Size, Port, and Auth", func() {
			obj.Spec.Size = 0
			obj.Spec.Port = 0
			obj.Spec.Auth = nil

			err := defaulter.Default(ctx, obj)
			Expect(err).To(BeNil())
			Expect(obj.Spec.Size).To(Equal(int32(1)))
			Expect(obj.Spec.Port).To(Equal(int32(8080)))
			Expect(*obj.Spec.Auth).To(BeFalse())
		})
	})

	Context("Validating Webhook", func() {
		It("should allow valid creation", func() {
			obj.Spec.Size = 1
			obj.Spec.Image = "nginx:latest"

			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(err).To(BeNil())
			Expect(warnings).To(BeEmpty())
		})

		It("should deny creation with invalid fields", func() {
			obj.Spec.Size = 0
			obj.Spec.Image = ""

			_, err := validator.ValidateCreate(ctx, obj)
			Expect(err).To(HaveOccurred())
		})

		It("should allow valid updates", func() {
			oldObj.Spec.Image = "nginx:old"
			obj.Spec.Image = "nginx:latest"

			warnings, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).To(BeNil())
			Expect(warnings).To(BeEmpty())
		})

		It("should deny invalid updates", func() {
			oldObj.Spec.Image = "nginx:old"
			obj.Spec.Image = ""

			_, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).To(HaveOccurred())
		})

		It("should allow deletion", func() {
			warnings, err := validator.ValidateDelete(ctx, obj)
			Expect(err).To(BeNil())
			Expect(warnings).To(BeEmpty())
		})
	})
})