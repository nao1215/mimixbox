#!/bin/bash
# @(#) libshell.sh is bash function library.
#
# Description :
#     libshell.sh is bash function library.
#     If you use this library, source this library in bash script.
#
#     [e.g. write down the following command in bash script(*.sh file)]
#      source <absolute path to this library>/libshell.sh
#

# errMsg() output message in red to stdout.
# arg1: message.
function errMsg() {
    local message="$1"
    echo -n -e "\033[31m\c"  # Escape sequence to make text color red
    echo "${message}" >&2
    echo -n -e "\033[m\c"  　# Escape sequence to restore font color
}


# warnMsg() output message in yellow to stdout.
# arg1: message.
function warnMsg() {
    local message="$1"
    echo -n -e "\033[33m\c"  # Escape sequence to make text color yellow
    echo "${message}"
    echo -n -e "\033[m\c"  　# Escape sequence to restore font color
}

# isRoot() check whether user is root or not.
function isRoot() {
    if [ ${EUID:-${UID}} != 0 ]; then
        echo "1"
        return
    fi
    echo "0"
}

# getAbsPath() get the absolute path where
# the script exists.
# arg1: script path(This mean "$0", itself)
function getAbsPath() {
	local abs_path=""
	local script_name="$1"

	abs_path=$(cd $(dirname $1); pwd)
	echo ${abs_path}
}


# upper() change from lower case to upper case.
# arg1: strings to be changed.
function upper() {
	local str="$1"
    echo -n ${str} | tr '[a-z]' '[A-Z]'
}

# lower() change from upper case to lower case.
# arg1: strings to be changed.
function lower() {
	local str="$1"
    echo -n ${str} | tr '[A-Z]' '[a-z]'
}
