#!/bin/zsh

wavsDir=$1
outputDir=$2
k=$3

#./make_test_data.sh ${wavsDir} ${outputDir}

gmmDir="${outputDir}/gmms"

if [[ ! -d ${gmmDir} ]]; then
    mkdir ${gmmDir}
fi

trainDir=$(echo ${outputDir}/train)
testDir=$(echo ${outputDir}/test)

echo "=====TRAIN ${k}======"
    go run cmd/train_emotions/main.go ${k} "${gmmDir}/gmm" -h ${trainDir}/happiness/* -s ${trainDir}/sadness/* -a ${trainDir}/anger/* -n ${trainDir}/neutral/* 

echo "=====TEST ${k}======"    
    go run cmd/test_emotion/main.go "${gmmDir}/gmm_k${k}" -h ${testDir}/happiness/* -s ${testDir}/sadness/* -a ${testDir}/anger/* -n ${testDir}/neutral/* > ${outputDir}/result_k${k}
