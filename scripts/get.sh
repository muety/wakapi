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
  latestJSON="$( eval "$http 'https://api.github.com/repos/$1/releases/latest'" 2>/dev/null )" || true

  versionNumber=''
  if ! echo "$latestJSON" | grep 'API rate limit exceeded' >/dev/null 2>&1 ; then
    if ! versionNumber="$( echo "$latestJSON" | grep -oEm1 '[0-9]+[.][0-9]+[.][0-9]+' - 2>/dev/null )" ; then
      versionNumber=''
    fi
  fi

  if [ "${versionNumber:-x}" = "x" ] ; then
    # Try to fallback to previous latest version detection method if curl is available
    if command -v curl >/dev/null 2>&1 ; then
      if finalUrl="$( curl "https://github.com/$1/releases/latest" -s -L -I -o /dev/null -w '%{url_effective}' 2>/dev/null )" ; then
        trimmedVers="${finalUrl##*v}"
        if [ "${trimmedVers:-x}" != "x" ] ; then
          echo "$trimmedVers"
          exit 0
        fi
      fi
    fi

    cat 1>&2 << 'EOA'
/=====================================\\
|     FAILED TO HTTP DOWNLOAD FILE     |
\\=====================================/

Uh oh! We couldn't download needed internet resources for you. Perhaps you are
 offline, your DNS servers are not set up properly, your internet plan doesn't
 include GitHub, or the GitHub servers are down?

EOA
    exit 1
  else
    echo "$versionNumber"
  fi
}

if [ "${GETWAKAPI_HTTP:-x}" != "x" ]; then
  http="$GETWAKAPI_HTTP"
elif command -v curl >/dev/null 2>&1 ; then
  http="curl -L"
elif command -v wget >/dev/null 2>&1 ; then
  http="wget -O-"
else
  cat 1>&2 << 'EOA'
/=====================================\\
|     COULD NOT FIND HTTP PROGRAM     |
\\=====================================/

Uh oh! We couldn't find either curl or wget installed on your system.

To continue with installation, you have two options:

A. Install either wget or curl on your system. You may need to run `hash -r`.

B. Define GETWAKAPI_HTTP to be a command (with arguments deliminated by spaces)
    that both follows HTTP redirects AND prints the fetched content to stdout.

For examples of option B, this script uses the below values for wget and curl:

  $ curl https://wakapi.dev/get | GETWAKAPI_HTTP="curl -L" sh

  $ wget -O- https://wakapi.dev/get | GETWAKAPI_HTTP="wget -O-" sh

EOA
  exit 1
fi

platform=''
machine=$(uname -m)

if [ "${GETWAKAPI_PLATFORM:-x}" != "x" ]; then
  platform="$GETWAKAPI_PLATFORM"
else
  case "$(uname -s | tr '[:upper:]' '[:lower:]')" in
    "linux")
      case "$machine" in
        "arm64"* | "aarch64"* ) platform='linux_arm64' ;;
        *"64") platform='linux_amd64' ;;
      esac
      ;;
    "darwin")
      case "$machine" in
        "arm64"* | "aarch64"* ) platform='darwin_arm64' ;;
        *"64") platform='darwin_amd64' ;;
      esac;;
    "msys"*|"cygwin"*|"mingw"*|*"_nt"*|"win"*)
      case "$machine" in
        "arm64"* | "aarch64"* ) platform='win_arm64' ;;
        *"64") platform='win_amd64' ;;
      esac
      ;;
  esac
fi

if [ "${platform:-x}" = "x" ]; then
  cat 1>&2 << 'EOM'
/=====================================\\
|      COULD NOT DETECT PLATFORM      |
\\=====================================/

Uh oh! We couldn't automatically detect your operating system. You can file a
bug here: https://github.com/muety/wakapi

To continue with installation, please choose from one of the following values:

- win_amd64
- darwin_amd64
- linux_amd64

Export your selection as the GETWAKAPI_PLATFORM environment variable, and then
re-run this script.

For example:

  $ curl https://getmic.ro | GETWAKAPI_PLATFORM=linux_amd64 sh

EOM
  exit 1
else
  echo "Detected platform: $platform"
fi

TAG=$(githubLatestTag muety/wakapi)

if command -v grep >/dev/null 2>&1 ; then
  if ! echo "v$TAG" | grep -E '^v[0-9]+[.][0-9]+[.][0-9]+$' >/dev/null 2>&1 ; then
      cat 1>&2 << 'EOM'
/=====================================\\
|         INVALID TAG RECIEVED         |
\\=====================================/

Uh oh! We recieved an invalid tag and cannot be sure that the tag will not break
 this script.

Please open an issue on GitHub at https://github.com/muety/wakapi with
 the invalid tag included:

EOM
    echo "> $TAG" 1>&2
    exit 1
  fi
fi

extension='zip'

echo "Latest Version: $TAG"
echo "Downloading https://github.com/muety/wakapi/releases/download/$TAG/wakapi_$platform.$extension"

eval "$http 'https://github.com/muety/wakapi/releases/download/$TAG/wakapi_$platform.$extension'" > "wakapi.$extension"

case "$extension" in
  "zip") unzip -j "wakapi.$extension" -d "wakapi-$TAG" ;;
  "tar.gz") tar -xvzf "wakapi.$extension" "wakapi-$TAG/wakapi" ;;
esac

mv wakapi-$TAG/* .

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
