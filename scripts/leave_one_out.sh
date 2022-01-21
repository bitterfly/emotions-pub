#!/bin/zsh

find_matches() {
    key="${1}"
    shift
    for dir in "${@}"; do
        find ${dir} -type f -name "${key}*" -exec readlink -f {} \;
    done | shuf
}



result_file="${1}"
a="${2}"

if [ ! -z "${3}" ]; then
    b="${3}"
fi

echo "" > ${result_file}


negative=$(find_matches "negative${keyword}*" "${a}" "${b}")
positive=$(find_matches "positive${keyword}*" "${a}" "${b}")
neutral=$(find_matches "neutral${keyword}*" "${a}" "${b}")

c=$(echo "${negative}" | wc -l)
len=$(perl -e "print int(${c} * 0.8)")
negative_train=$(echo "${negative}" | head -n "${len}")
negative_test=$(echo "${negative}" | tail -n +$((len+1)))

c=$(echo "${positive}" | wc -l)
len=$(perl -e "print int(${c} * 0.8)")
positive_train=$(echo "${positive}" | head -n "${len}")
positive_test=$(echo "${positive}" | tail -n +$((len+1)))

c=$(echo "${neutral}" | wc -l)
len=$(perl -e "print int(${c} * 0.8)")
neutral_train=$(echo "${neutral}" | head -n "${len}")
neutral_test=$(echo "${neutral}" | tail -n +$((len+1)))

if [[ $(comm -12 <(echo "${negative_train}" | sort) <(echo "${negative_test}" | sort) | wc -l ) > 0 ]]; then
    echo "Negative train and negative test have common elements:"
    comm -12 <(echo "${negative_train}" | sort) <(echo "${negative_test}" | sort)
    exit
fi

if [[ $(comm -12 <(echo "${positive_train}" | sort) <(echo "${positive_test}" | sort) | wc -l ) > 0 ]]; then
    echo "Positive train and positive test have common elements:"
    comm -12 <(echo "${positive_train}" | sort) <(echo "${positive_test}" | sort)
    exit
fi

if [[ $(comm -12 <(echo "${neutral_train}" | sort) <(echo "${neutral_test}" | sort) | wc -l ) > 0 ]]; then
    echo "Neutral train and neutral test have common elements:"
    comm -12 <(echo "${neutral_train}" | sort) <(echo "${neutral_test}" | sort)
    exit
fi

./make_input_file.sh --eeg-positive $(echo ${positive_test}) --eeg-negative $(echo ${negative_test}) --eeg-neutral $(echo ${neutral_test}) > "/tmp/eeg-test"
./make_input_file.sh --eeg-positive $(echo ${positive_train}) --eeg-negative $(echo ${negative_train}) --eeg-neutral $(echo ${neutral_train}) > "/tmp/eeg-train"

echo -e "Whole file\teeg-negative\teeg-neutral\teeg-possitive" >> "${result_file}"
train_eeg 1 /tmp/foo "/tmp/eeg-train"
test_eeg 1 /tmp/foo "/tmp/eeg-test" 2>/tmp/log >> "${result_file}"
echo "" >> ${result_file}

for dur in `seq 2800 200 5000`; do 
    echo -e "${dur}\teeg-negative\teeg-neutral\teeg-possitive" >> ${result_file}
    
    train_eeg "${dur}" /tmp/foo "/tmp/eeg-train"
    test_eeg "${dur}" /tmp/foo "/tmp/eeg-test" 2>/tmp/log >> "${result_file}"
    echo "" >> "${result_file}"
done

rm "/tmp/foo"
# rm "/tmp/eeg-train"
# rm "/tmp/eeg-test"