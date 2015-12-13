#!/bin/bash
hosts=(10.10.0.2 10.10.0.3 10.10.0.4 10.10.0.5 10.10.0.13 10.10.0.6 10.10.0.10 10.10.0.9 10.10.0.14 10.10.0.7 10.10.0.11)

for host in ${hosts[@]}; do
    scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
        sysward root@$host:/opt/sysward/bin/ &
done

wait
