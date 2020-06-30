#!/bin/bash
#
# Copyright 2020 Google LLC
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
#
# This script runs customization scripts on a COS VM instance. It pulls
# source from GCS and executes it.

set -o errexit
set -o pipefail
set -o nounset

trap 'fatal exiting due to errors' EXIT

mkdir -p /etc/systemd/system/systemd-journald.service.d
cat > /etc/systemd/system/systemd-journald.service.d/override.conf<<EOF
[Service]
Restart=no
EOF

/usr/bin/systemctl daemon-reload
/usr/bin/systemctl stop systemd-journald.socket
/usr/bin/systemctl stop systemd-journald-dev-log.socket
/usr/bin/systemctl stop systemd-journald-audit.socket
/usr/bin/systemctl stop syslog.socket
/usr/bin/systemctl stop systemd-journald.service


cat > /etc/systemd/system/last-run.service<<EOF
[Unit]
Description=Run after everything unmounted
DefaultDependencies=false
Conflicts=shutdown.target
Before=shutdown.target multi-user.target mnt-stateful_partition.mount var.mount mnt-disks.mount var-lib-docker.mount var-lib-toolbox.mount usr-share-oem.mount
After=tmp-last.mount

[Service]
Type=oneshot
RemainAfterExit=true
ExecStart=/bin/true
ExecStop=/bin/sh -c 'exec -c /tmp/extend-oem.bin /dev/sda 1 8 {{size}}'
TimeoutStopSec=600
EOF

systemctl start last-run.service