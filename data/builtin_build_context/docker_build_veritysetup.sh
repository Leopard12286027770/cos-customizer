#!/bin/bash
#
# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
# Build docker image with veritysetup. 
# The image veritysetup depend on ubuntu image, 
# so the ubuntu image cannot be deleted directly.
# In order to delete ubuntu image to save disk space, 
# we need to save the image, delete docker images,
# and then load it from the saved file.
sudo docker build -t veritysetup .
sudo docker save -o veritysetup.img veritysetup
sudo docker image prune -af
sudo docker load -i veritysetup.img
sudo rm -f veritysetup.img