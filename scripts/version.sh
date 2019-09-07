#!/bin/bash

function require_clean_work_tree () {
    # Update the index
    git update-index -q --ignore-submodules --refresh
    err=0

    # Disallow unstagged changes in the working tree
    if ! git diff-files --quiet --ignore-submodules --
    then
        echo >&2 "cannot $1: you have unstaged changes."
        git diff-files --name-status -r --ignore-submodules -- >&2
        err=1
    fi

    # Disallow uncommitted changes in the index
    if ! git diff-index --cached --quiet HEAD --ignore-submodules --
    then
        echo >&2 "cannot $1: your index contains uncommitted changes."
        git diff-index --cached --name-status -r --ignore-submodules HEAD -- >&2
        err=1
    fi

    if [[ ${err} = 1 ]]
    then
        echo >&2 "Please commit or stash them."
        exit 1
    fi
}

function menu(){
    clear
    printf "Select what you want to update: \n"
    printf "1 - Major update\n"
    printf "2 - Minor update\n"
    printf "3 - Patch update\n"
    printf "4 - Exit\n"
    read -r selection


    case "$selection" in
        1) printf "Major updates......\n"
            NEW_VERSION=$(git tag | sed 's/\(.*v\)\([0-9]*\)\.\([0-9]*\)\.\([0-9]*\)/\2;\3;\4;\1/g' | sort  -t';' -k 1,1n  -k 2,2n -k 3,3n | tail -n 1  | awk -F';' '{printf "%s%d.%d.%d", $4, ($1+1),0,0 }')
        ;;
        2) printf "Run Minor update.........\n"
            NEW_VERSION=$(git tag | sed 's/\(.*v\)\([0-9]*\)\.\([0-9]*\)\.\([0-9]*\)/\2;\3;\4;\1/g' | sort  -t';' -k 1,1n  -k 2,2n -k 3,3n | tail -n 1  | awk -F';' '{printf "%s%d.%d.%d", $4, $1,($2+1),0 }')
        ;;
        3) printf "Patch update.........\n"
            NEW_VERSION=$(git tag | sed 's/\(.*v\)\([0-9]*\)\.\([0-9]*\)\.\([0-9]*\)/\2;\3;\4;\1/g' | sort  -t';' -k 1,1n  -k 2,2n -k 3,3n | tail -n 1  | awk -F';' '{printf "%s%d.%d.%d", $4, $1,$2,($3 + 1) }')
        ;;
        4) printf "Exit................................\n"
            exit 1
        ;;
        *) clear
            printf "Incorrect selection. Try again\n"
            menu
        ;;
    esac


}

## Check if git is clean
require_clean_work_tree "create new version"

git pull

## Sem ver update menu
menu

if [[ "${NEW_VERSION}" = "" ]]; then
    NEW_VERSION="v1.0.0"
fi

echo ${NEW_VERSION}

message="version ${NEW_VERSION}"
ADD="version"
read -r  -n 1 -p "y?:" userok
echo ""
if [[ "$userok" = "y" ]]; then
    read -r -n 1 -p "Update commit message?: y/n" userok
    echo ""
    if [[ "$userok" = "y" ]]; then
        read -r -p "Message: " message
        ADD="."
        echo ""
    fi
    echo ${NEW_VERSION} > version && git add ${ADD} && git commit -m "$message"&& git tag -a ${NEW_VERSION} -m ${NEW_VERSION} && git push --tags && git push
fi
echo
