# Copyright 2017 Google Inc. All rights reserved.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to writing, software distributed
# under the License is distributed on a "AS IS" BASIS, WITHOUT WARRANTIES OR
# CONDITIONS OF ANY KIND, either express or implied.
#
# See the License for the specific language governing permissions and
# limitations under the License.

INSTANCE:="friendfinder"
ZONE:="us-central1-f"
USER:="ianhoegen"

friendfinder:
	GOOS=linux go build twitter/twitter.go -o friendfinder

clean:
	rm -f friendfinder

instance:
	gcloud compute instances describe --zone $(ZONE) $(INSTANCE) &> /dev/null || \
	gcloud compute instances create $(INSTANCE) \
		--zone $(ZONE) --machine-type "f1-micro" \
		--image "debian-8-jessie-v20170523" --image-project "debian-cloud";

deploy: instance friendfinder
	gcloud compute scp --zone $(ZONE) friendfinder friendfinder.service $(USER)@$(INSTANCE):~
	gcloud compute ssh --zone $(ZONE) $(USER)@$(INSTANCE) --command \
		"sudo mv ~/friendfinder.service /etc/systemd/system/"
	gcloud compute ssh --zone $(ZONE) $(USER)@$(INSTANCE) --command \
		"sudo systemctl enable friendfinder && sudo systemctl start friendfinder"