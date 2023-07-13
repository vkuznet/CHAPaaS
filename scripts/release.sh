#!/bin/bash
#
# check user input
if [ $# -ne 3 ]; then
    echo "Not enough arguments, usage: release.sh <user-repo> <release-tag> <release-notes>"
    exit 1;
fi

dir=$1   # e.g /path/CHAPUsers
tag=$2   # e.g. v0.0.0
notes=$3 # e.g. "some release notes"
token=$CHAPUSERS_TOKEN

echo "Create release:"
echo "Directory     : $dir"
echo "Tag           : $tag"
echo "Notes         : $notes"
payload=$(printf '{"tag_name": "%s","target_commitish": "main","name": "Auto-generated release %s","body": "%s","draft": false,"prerelease": false}' $tag $tag "$notes")
echo "payload       : $payload"

if [ -z "$token" ]; then
    echo "Please define CHAPUSERS_TOKEN environment with proper github access token for CHAPUsers repo"
    exit 1;
fi

repo="CHAPUsers"
owner="CHESSComputing"

# see: https://docs.github.com/en/rest/releases/releases?apiVersion=2022-11-28
curl -k -s -L \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer $token"\
  -H "X-GitHub-Api-Version: 2022-11-28" \
  -d "$payload" \
  https://api.github.com/repos/$owner/$repo/releases