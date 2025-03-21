/*
Copyright 2018 The Kubernetes Authors.

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

package metadata

type fakeServiceManager struct{}

var _ MetadataService = &fakeServiceManager{}

var (
	FakeMachineType     = "n1-standard-1"
	FakeZone            = "country-region-zone"
	FakeProject         = "test-project"
	FakeName            = "test-name"
	FakeClusterLocation = "us-central1"
	FakeClusterName     = "test-cluster"
)

func NewFakeService() MetadataService {
	return &fakeServiceManager{}
}

func (manager *fakeServiceManager) GetZone() string {
	return FakeZone
}

func (manager *fakeServiceManager) GetProject() string {
	return FakeProject
}

func (manager *fakeServiceManager) GetName() string {
	return FakeName
}

func (manager *fakeServiceManager) GetMachineType() string {
	return FakeMachineType
}

func (manager *fakeServiceManager) GetClusterLocation() string {
	return FakeClusterLocation
}

func (manager *fakeServiceManager) GetClusterName() string {
	return FakeClusterName
}

func SetMachineType(s string) {
	FakeMachineType = s
}

func SetZone(s string) {
	FakeZone = s
}

func SetName(s string) {
	FakeName = s
}

func SetClusterLocation(s string) {
	FakeClusterLocation = s
}

func SetClusterName(s string) {
	FakeClusterName = s
}
