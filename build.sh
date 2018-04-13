#!/bin/sh
filepath=$(cd `dirname $0`; pwd)
version=$(grep 'Version' ${filepath}/src/main/main.go|grep -o '=.*'|grep -o '[0-9.]*')
projectname="asset-monitor"
tar -zcvf ${filepath}/${projectname}-v${version}.tar.gz -C ${filepath} etc bin doc README.md