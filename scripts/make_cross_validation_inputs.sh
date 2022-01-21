#!/bin/zsh

if [ $# -le 3 ]; then
    echo "usage: <dir> <output-dir> <batch-num> [<ext>]"
    exit 1
fi

dir="${1}"
output_dir="${2}"
batchnum="${3}"
ext="${4}"

if [ -z ${ext} ]; then
    ext="wav"
fi

if [ ! -d ${output_dir} ]; then
    mkdir ${output_dir}
fi

rm -rf ${output_dir}/*

files=$(find "${dir}" -type f -name "*.${ext}")

anger=$(echo  ${files} | grep "anger")
happiness=$(echo  ${files} | grep "happiness")
sadness=$(echo  ${files} | grep "sadness")
neutral=$(echo  ${files} | grep "neutral")
echo -e "anger: $(echo ${anger} |wc -l)"
echo -e "happiness: $(echo ${happiness} |wc -l)"
echo -e "sadness: $(echo ${sadness} |wc -l)"
echo -e "neutral: $(echo ${neutral} |wc -l)"

anger_batch=$(echo ${anger} | wc -l | awk -v batchnum=${batchnum} '{print int($0/batchnum)}')
happiness_batch=$(echo ${happiness} | wc -l | awk  -v batchnum=${batchnum} '{print int($0/batchnum)}')
sadness_batch=$(echo ${sadness} | wc -l | awk  -v batchnum=${batchnum} '{print int($0/batchnum)}')
neutral_batch=$(echo ${neutral} | wc -l | awk  -v batchnum=${batchnum} '{print int($0/batchnum)}')

for i in $(seq 1 ${batchnum}); do
    batch=$(echo ${anger} | shuf -n ${anger_batch})
    paste <(yes "anger" | head -n ${anger_batch}) <(echo ${batch}) >> ${output_dir}/batch_${i}.txt
    anger=$(comm -23 <(echo ${anger}|sort) <(echo ${batch} | sort))

    batch=$(echo ${happiness} | shuf -n ${happiness_batch})
    paste <(yes "happiness" | head -n ${happiness_batch}) <(echo ${batch}) >> ${output_dir}/batch_${i}.txt
    happiness=$(comm -23 <(echo ${happiness}|sort) <(echo ${batch} | sort))

    batch=$(echo ${sadness} | shuf -n ${sadness_batch})
    paste <(yes "sadness" | head -n ${sadness_batch}) <(echo ${batch}) >> ${output_dir}/batch_${i}.txt
    sadness=$(comm -23 <(echo ${sadness}|sort) <(echo ${batch} | sort))

    batch=$(echo ${neutral} | shuf -n ${neutral_batch})
    paste <(yes "neutral" | head -n ${neutral_batch}) <(echo ${batch}) >> ${output_dir}/batch_${i}.txt
    neutral=$(comm -23 <(echo ${neutral}|sort) <(echo ${batch} | sort))

    sort -o ${output_dir}/batch_${i}.txt ${output_dir}/batch_${i}.txt
done
