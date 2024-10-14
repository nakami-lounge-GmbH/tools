#!/bin/bash
# this script deletes all directories under the version subdir, except the one which is linked to the actual binary
# please call the script form the api directory (w.g. /srv/vkb/em/api)
# param $1 must be the name of the binary

if [ $# -eq 0 ]
  then
    echo "Please specify the name of the binary, e.g. vkb.2vm07-go"
    exit -1
fi

subdir=$(ls -alF $1 |awk '{print $11}'|awk -F "/" '{ print $(NF-1) }')
echo "Not deleting <$subdir>"
find ./versions -mindepth 1 ! -regex "^./versions/$subdir\(/.*\)?" -print
