#!/bin/bash

ANY_FAILED=0
check() {
	if RES="$(dig +noall +answer @127.0.0.1 -p 5353 $1)"
	then
		if [[ $(echo $RES | awk '{print $1}') != $1. ]]
		then
			ANY_FAILED=1
			echo -e "\e[1;31mFailed: \e[0;0mcheck \e[0;37m$1 $2\e[0;0m"
			echo -e "        expected answer name to be \e[0;34m'$1.'\e[0;0m, got \e[0;34m'$(echo $RES | awk '{print $1}')'\e[0;0m"
			return 1
		fi
		if [[ $(echo $RES | awk '{print $5}') != $2 ]]
		then
			ANY_FAILED=1
			echo -e "\e[1;31mFailed: \e[0;0mcheck \e[0;37m$1 $2\e[0;0m"
			echo -e "        expected answer IP to be \e[0;34m'$2'\e[0;0m, got \e[0;34m'$(echo $RES | awk '{print $5}')'\e[0;0m"
			return 1
		fi
		return 0
	fi
	ANY_FAILED=1
	echo -e "\e[1;31mFailed: \e[0;0mcheck \e[0;37m$1 $2\e[0;0m"
	echo  '        dig failed'
	return 1
}

../coredns -conf Corefile -p 5353 >/dev/null 2>/dev/null &
sleep 1
COREDNS_PID="$!"

check example.com 127.0.0.1
check bar.foo.example.com 127.0.0.1
check google-analytics.com 0.0.0.0

kill "$COREDNS_PID"
if [[ $ANY_FAILED == 0 ]]
then
	echo -e '\e[1;32m[all tests passed]\e[0;0m Ready to roll 🚀'
else
	echo -e '\e[1;31m[some tests failed]\e[0;0m Go fix it 💩'
fi
exit $ANY_FAILED
