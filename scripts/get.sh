#!/bin/sh

# This script installs Wakapi.
#
# Quick install: `curl https://wakapi.dev/get | bash`
#
# This script will install Wakapi to the directory you're in. To install
# somewhere else (e.g. /usr/local/bin), cd there and make sure you can write to
# that directory, e.g. `cd /usr/local/bin; curl https://wakapi.dev/get | sudo bash`
#
# Acknowledgments:
#   - Micro Editor for this script: https://micro-editor.github.io/
#   - ASCII art courtesy of figlet: http://www.figlet.org/

set -e -u

githubLatestTag() {
  finalUrl=$(curl "https://github.com/$1/releases/latest" -s -L -I -o /dev/null -w '%{url_effective}')
  printf "%s\n" "${finalUrl##*/}"
}

platform=''
machine=$(uname -m) # currently, Wakapi builds are only available for AMD64 anyway

if [ "${GETWAKAPI_PLATFORM:-x}" != "x" ]; then
  platform="$GETWAKAPI_PLATFORM"
else
  case "$(uname -s | tr '[:upper:]' '[:lower:]')" in
    "linux") platform='linux_amd64' ;;
    "msys"*|"cygwin"*|"mingw"*|*"_nt"*|"win"*) platform='win_amd64' ;;
  esac
fi

if [ "x$platform" = "x" ]; then
  cat << 'EOM'
/=====================================\\
|      COULD NOT DETECT PLATFORM      |
\\=====================================/

Uh oh! We couldn't automatically detect your operating system. You can file a
bug here: https://github.com/muety/wakapi
EOM
  exit 1
else
  printf "Detected platform: %s\n" "$platform"
fi

TAG=$(githubLatestTag muety/wakapi)

printf "Tag: %s" "$TAG"

extension='zip'

printf "Latest Version: %s\n" "$TAG"
printf "Downloading https://github.com/muety/wakapi/releases/download/%s/wakapi_%s.%s\n" "$TAG" "$platform" "$extension"

curl -L "https://github.com/muety/wakapi/releases/download/$TAG/wakapi_$platform.$extension" > "wakapi.$extension"

case "$extension" in
  "zip") unzip -j "wakapi.$extension" -d "wakapi-$TAG" ;;
  "tar.gz") tar -xvzf "wakapi.$extension" "wakapi-$TAG/wakapi" ;;
esac

mv "wakapi-$TAG/wakapi" ./wakapi
mv "wakapi-$TAG/config.yml" ./config.yml

rm "wakapi.$extension"
rm -rf "wakapi-$TAG"

cat <<-'EOM'

__        __    _               _
\ \      / /_ _| | ____ _ _ __ (_)
 \ \ /\ / / _` | |/ / _` | '_ \| |
  \ V  V / (_| |   < (_| | |_) | |
   \_/\_/ \__,_|_|\_\__,_| .__/|_|
                         |_|

Wakapi has been downloaded to the current directory.
You can run it with:

./wakapi

For further instructions see https://github.com/muety/wakapi

EOM
