#!/bin/zsh

find_matches() {
    key="${1}"
    shift
    for dir in "${@}"; do
        find ${dir} -type f -name "${key}" -exec readlink -f {} \;
    done | shuf
}

train_file="${1}"
shift
test_file="${1}"
shift

negative=$(find_matches "negative*" "${@}")
positive=$(find_matches "positive*" "${@}")
neutral=$(find_matches "neutral*" "${@}")

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

./make_input_file.sh --positive $(echo ${positive_train}) --negative $(echo ${negative_train}) --neutral $(echo ${neutral_train}) > "${train_file}"
./make_input_file.sh --positive $(echo ${positive_test}) --negative $(echo ${negative_test}) --neutral $(echo ${neutral_test}) > "${test_file}" 
