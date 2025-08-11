#
# Copyright Flow Foundation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

.PHONY: ci
ci:
	# test with coverage
	go test -coverprofile=coverage.txt -covermode=atomic -parallel 8 -race -coverpkg $(COVERPKGS) ./...
	# remove coverage of empty functions from report
	sed -i -e 's/^.* 0 0$$//' coverage.txt

.PHONY: test
test:
	(go test -parallel 8 ./...)
