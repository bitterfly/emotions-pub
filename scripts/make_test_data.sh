#!/bin/zsh

function split {
    emotion=$1

    eTrainDir=$(echo ${2}/$(basename ${emotion}))
    eTestDir=$(echo ${3}/$(basename ${emotion}))

    if [[ ! -d  ${eTrainDir} ]]; then
        mkdir ${eTrainDir}
    fi

    if [[ ! -d  ${eTestDir} ]]; then
        mkdir ${eTestDir}
    fi


    shufFiles=$(find ${emotion} -mindepth 1 -maxdepth 1 -type f -name "*.wav" | shuf)
    filesNum=$(echo ${shufFiles} | wc -l)


    eightyPercent=$(((80 * filesNum)/100 ))

    i=1
    while read file; do
        if [ ${i} -le ${eightyPercent} ]; then
            cp -v ${file} "${eTrainDir}/"
        else
            cp -v ${file} "${eTestDir}/"
        fi

        i=$((i+=1))

    done < <(echo ${shufFiles})
}

wavsDir=$(realpath $1)
outputDir=$(realpath $2)

if [[ ! -d ${outputDir} ]]; then
    mkdir ${outputDir}
fi

trainDir="${outputDir}/train/"
testDir="${outputDir}/test/"

if [[ ! -d ${trainDir} ]]; then
    mkdir ${trainDir}
fi

if [[ ! -d ${testDir} ]]; then
    mkdir ${testDir}
fi

angerDir=$(find ${wavsDir} -maxdepth 1 -type d -name "*anger*" -exec realpath {} \;)
happinessDir=$(find ${wavsDir} -maxdepth 1 -type d -name "*happiness*" -exec realpath {} \;)
sadnessDir=$(find ${wavsDir} -maxdepth 1 -type d -name "*sadness*" -exec realpath {} \;)
neutralDir=$(find ${wavsDir} -maxdepth 1 -type d -name "*neutral*" -exec realpath {} \;)

echo "anger: ${angerDir}"
echo "happiness: ${happinessDir}"
echo "sadness: ${sadnessDir}"
echo "neutral: ${neutralDir}"

echo "train: ${trainDir}"
echo "test: ${testDir}"

split ${angerDir} ${trainDir} ${testDir}
split ${happinessDir} ${trainDir} ${testDir}
split ${neutralDir} ${trainDir} ${testDir}
split ${sadnessDir} ${trainDir} ${testDir}
