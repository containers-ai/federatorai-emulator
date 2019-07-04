#!/bin/sh
#
AIHOME="/opt/alameda/federatorai-agent"
#
ns=$1
pod=$2
dest_dir=$3

[ "${dest_dir}" = "" ] && show_usage

show_usage()
{
    echo ""
    echo "    Usage: $0 [namespace] [pod] [dest_dir]"
    echo ""
    exit 1
}

#
# main
#
# version.txt
kubectl -n ${ns} cp ${pod}:${AIHOME}/etc/version.txt ${dest_dir}/version.txt

# logs
kubectl -n ${ns} exec ${pod} -- find /var/log/alameda -type f \
  | while read fn; do
        kubectl -n ${ns} cp ${pod}:${fn} ${dest_dir}
    done

exit 0
