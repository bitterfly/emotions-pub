#!/bin/zsh

train_directory=${1}
test_directory=${2}
result_file=${3}

echo "" > ${result_file}

negative_train=$(find ${train_directory} -type f -name "negative*" -exec readlink -f {} \;)
positive_train=$(find ${train_directory} -type f -name "positive*" -exec readlink -f {} \;)
neutral_train=$(find ${train_directory} -type f -name "neutral*" -exec readlink -f {} \;)

negative_test=$(find ${test_directory} -type f -name "negative*" -exec readlink -f {} \;)
positive_test=$(find ${test_directory} -type f -name "positive*" -exec readlink -f {} \;)
neutral_test=$(find ${test_directory} -type f -name "neutral*" -exec readlink -f {} \;)

echo -e "Whole file\teeg-negative\teeg-neutral\teeg-possitive" >> ${result_file}
train_eeg 1 /tmp/foo --eeg-positive $(echo ${positive_train}) --eeg-negative $(echo ${negative_train}) --eeg-neutral $(echo ${neutral_train})
test_eeg 1 /tmp/foo --eeg-positive $(echo ${positive_test}) --eeg-negative $(echo ${negative_test}) --eeg-neutral $(echo ${neutral_test}) 2>/dev/null >> ${result_file}
echo "" >> ${result_file}

for dur in `seq 200 200 5000`; do 
    echo -e "${dur}\teeg-negative\teeg-neutral\teeg-possitive" >> ${result_file}
    
    train_eeg ${dur} /tmp/foo --eeg-positive $(echo ${positive_train}) --eeg-negative $(echo ${negative_train}) --eeg-neutral $(echo ${neutral_train})
    test_eeg ${dur} /tmp/foo --eeg-positive $(echo ${positive_test}) --eeg-negative $(echo ${negative_test}) --eeg-neutral $(echo ${neutral_test}) 2>/tmp/log >> ${result_file}
    echo "" >> ${result_file}
done