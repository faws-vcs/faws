#!/usr/bin/env bash
set -e

package_name="faws-vcs/faws"
program_name="faws"

get_latest_release() {
  curl -s "https://api.github.com/repos/${package_name}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
}

latest_version=$(get_latest_release)

temporary_directory=$(mktemp -d)

release_os=$(uname -s)
release_arch=$(uname -m)
release_filename="${program_name}_${release_os}_${release_arch}.tar.gz"

echo "installing ${program_name} ${latest_version} for ${release_os}/${release_arch}"

# Download GoReleaser package tarball
package_release_url="https://github.com/${package_name}/releases/download/${latest_version}/${release_filename}"
curl -sLo ${temporary_directory}/${release_filename} $package_release_url

# Extract GoReleaser package tarball
extract_directory="${temporary_directory}/${program_name}_${release_os}_${release_arch}"
mkdir $extract_directory
tar -xzf "${temporary_directory}/${release_filename}" -C $extract_directory

# Install binary
sudo mv "${extract_directory}/${program_name}" "/usr/local/bin/${program_name}"

# Setup autocompletion
setup_bash_completion() {
    echo "setting up bash completion"
    sudo mkdir -p /etc/bash_completion.d
    /usr/local/bin/faws completion bash | sudo tee /etc/bash_completion.d/faws.sh > /dev/null
    sudo chmod 0644 /etc/bash_completion.d/faws.sh
    source /etc/bash_completion.d/faws.sh
}

setup_zsh_completion() {
    echo "zsh completion not yet implemented :("
}

case "$SHELL" in
    */bash) setup_bash_completion ;;
    */zsh)  setup_zsh_completion ;;
esac

echo "installed ${program_name} ${latest_version}"

# Remove GoReleaser package
echo "removing temporary directory"
rm -rf $temporary_directory