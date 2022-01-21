#!/bin/zsh
train_executable="${1}"
test_executable="${2}"
batch_files_dir="${3}"
gmm_models_dir="${4}"
result_dir="${5}"
k="${6}"
type=${7}
feature_type=${8}

if [ $# -le 5 ]; then
    echo "usage: <train-executable> <test-executable> <batch-dir> <model-dir> <result-dir> [<k> <type> <feature-type>]"
    echo "concat_classifier: <train> <test> <batch> <model> <result> 3"
    echo "speech_classifier: <train> <test> <batch> <model> <result> k"
    echo "eeg_classifier: <train> <test> <batch> <model> <result> 3 eeg [de/ge]"
    exit 1
fi

batch_files=$(find ${batch_files_dir} -type f)
for file in $(find ${batch_files_dir} -type f | sort); do
    if [[ ${type} == "eeg" ]];then
        echo "EEG"
        echo Training ${file}         
        "${train_executable}" "gmm" "${feature_type}" 0 ${gmm_models_dir}/gmm_$(basename ${file%.txt}) <(cat $(comm -23 <(echo ${batch_files} | sort) <(echo ${file})))
        echo Testing ${file} 
        "${test_executable}" "gmm" "${feature_type}" 0 ${gmm_models_dir}/gmm_$(basename ${file%.txt}) <(cat ${file}) > ${result_dir}/result_$(basename ${file%.txt}).res 2> ${result_dir}/result_$(basename ${file%.txt}).err 
    else 
        echo ${type}
        echo Training ${file}         
        "${train_executable}" ${k} ${gmm_models_dir}/gmm_$(basename ${file%.txt}) <(cat $(comm -23 <(echo ${batch_files} | sort) <(echo ${file})))
        echo Testing ${file} 
        "${test_executable}" ${gmm_models_dir}/gmm_$(basename ${file%.txt})_k${k} <(cat ${file}) > ${result_dir}/result_$(basename ${file%.txt})_k${k}
    fi
done
