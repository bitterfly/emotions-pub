#!/bin/zsh

recursiveDirname() {
	if [[ "${2}" == "1" ]]; then
		dirname "${1}"
	else
		dirname $(recursiveDirname  "${1}" $((${2} - 1)))
	fi
}

find_eeg(){
	eeg_file=$(basename "${1%wav}csv")
	find $(recursiveDirname "${1}" "3") -name "${eeg_file}"
}

format() {
	if [ -z "${3}" ]; then
		echo -e "${1}\t${2}"
	else
		echo -e "${1}\t${2}\t"$(find_eeg "${2}")
	fi
}

lastTag=""
eeg=""
while [[ $# > 0 ]]; do
	case "$1" in
		    "--eeg")
				eeg="1"
	            ;;
	        "--happiness")
	            lastTag="happiness"
	            ;;
		    "--sadness")
	            lastTag="sadness"
	            ;;
	
		    "--anger")
	            lastTag="anger"
	            ;;

			"--neutral")
	            lastTag="neutral"
	            ;;

			"--positive")
	            lastTag="positive"
	            ;;

			"--negative")
	            lastTag="negative"
	            ;;

			"--eeg-neutral")
	            lastTag="eeg-neutral"
	            ;;

			"--eeg-positive")
	            lastTag="eeg-positive"
	            ;;

			"--eeg-negative")
	            lastTag="eeg-negative"
	            ;;
	        *)	
				format "${lastTag}" "${1}" "${eeg}"
	esac
	shift
done