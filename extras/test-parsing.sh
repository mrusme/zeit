#!/usr/bin/env bash

############################################################
# Help                                                     #
############################################################
Help()
{
   echo "Run Parsing test suite for zeit"
   echo
   echo "Syntax: test-parsing.sh [-t ZEIT_DB_PATH] [-c CMD]"
   echo "options:"
   echo "h     Print this Help."
   echo "t     overwrite default ZEIT_DB Path for test (default: /tmp/zeit_test_parsing.db"
   echo "c     define command to test (default: zeit from Path)"
   echo "v     Verbose"
   echo
}

############################################################
# Process the input options. Add options as needed.        #
############################################################
# Get the options
while getopts ":t:c:" option; do
   case $option in
      h) # display Help
         Help
         exit;;
      t) # Overwrite tmp path  for test DB
         INPUT_PATH=$OPTARG;;
      c) # Set command to test
         CMD=$OPTARG;;
     \?) # Invalid option
         echo "Error: Invalid option"
         exit;;
   esac
done


if [[ -z $INPUT_PATH ]]; then
	DB_PATH=/tmp/zeit_test_parsing.db
else
  if [[ -d $INPUT_PATH ]]; then
	  DB_PATH="${INPUT_PATH}/zeit_test_parsing.db"
  else
		echo "ERROR: The Path entered for -t is either not existing or not a directory. Valid input is only an existing directory"
		exit 1
  fi
fi

if [[ -f $DB_PATH ]]; then
	rm $DB_PATH
fi

echo "PATH: $DB_PATH"

if [[ -z $CMD ]]; then
	CMD=$(command -v -- zeit)
else
	if [[ -z $CMD ]]; then
  	echo "ERROR: No Executable found to test, zeit not in path and set with -c"
		exit 1
	fi
fi


echo "CMD: $CMD"

declare -a tests
tests+=('1;10:00;11:00')
tests+=('2;01:00pm;02:00pm')
tests+=('3;-04:00;-03:00')
tests+=('4;-02:00;+01:00')
tests+=('5;2023-09-11 10:00 +0300;2023-09-11 12:00 +0400')
tests+=('6;2023-09-11 20:00;2023-09-11 21:00')
tests+=('7;2023-09-11T20:00;2023-09-11T21:00')
tests+=('8;2023-09-11T20:00:00;2023-09-11T21:00:00')
tests+=('9;2023-09-11T20:00:00+03:00;2023-09-11T21:00:00+03:00')
tests+=('10;01.04.2025 10:00;01.04.2025 12:00')
tests+=('11;25.05. 10:00;25.05. 12:00')
tests+=('12;04-01 10:00;04-01 12:00')
tests+=('13;01.04. 10:00;01.04. 12:00')
tests+=('14;01.04 10:00;01.04 12:00') # Will not work but parse to today without dot after month
tests+=('15;1 hour ago;in 2 hours')

for ((i = 0; i < ${#tests[@]}; i++))
do
  test_line=${tests[$i]}
  # echo "LINE: $test_line)"

  mapfile -td \; test < <(printf "%s\0" "$test_line")

  # echo ${test[1]}
  # echo ${test[2]}

  echo "$CMD track -p "TESTS" -t "Zeit-Test ${test[0]}" -b ${test[1]} -s ${test[2]}"
  $CMD track -p "TESTS" -t "Zeit-Test ${test[0]}" -b "${test[1]}" -s "${test[2]}"
  $CMD list | grep "Zeit-Test ${test[0]}"
  echo -e "\n"
done
