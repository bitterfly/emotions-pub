#!/bin/zsh
train_executable="${1}"
test_executable="${2}"
batch_files_dir="${3}"

speech_gmm_dir="${4}"
eeg_gmm_dir="${5}"

gmm_models_dir="${6}"
result_dir="${7}"
 
ks="${8}"
ke="${9}"

if [ $# -le 7 ]; then
    echo "usage: <train-executable> <test-executable> <batch-dir> <speech-gmm-dir> <eeg-gmm-dir> <result-model-dir> <result-dir> <k>"
    exit 1
fi

batch_files=$(find ${batch_files_dir} -type f)
for file in $(find ${batch_files_dir} -type f | sort); do
    "${train_executable}" "${speech_gmm_dir}/gmm_$(basename ${file%.txt})_k${ks}" ${eeg_gmm_dir}/gmm_$(basename ${file%.txt})  ${gmm_models_dir}/gmm_$(basename ${file%.txt}) <(cat $(comm -23 <(echo ${batch_files} | sort) <(echo ${file})))

    echo Testing ${file} 
    "${test_executable}" ${gmm_models_dir}/gmm_$(basename ${file%.txt})_speech ${gmm_models_dir}/gmm_$(basename ${file%.txt})_eeg <(cat ${file}) > ${result_dir}/result_$(basename ${file%.txt}).res 2> ${result_dir}/result_$(basename ${file%.txt}).err
done
