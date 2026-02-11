import sys
from typing import List

PATCH = 'patch'
MINOR = 'minor'
MAJOR = 'major'


class CliException(Exception):
    ...


def generate_version(current_version: str, patch: bool, minor: bool, major: bool) -> str:
    split_current_version = current_version.split('.')
    major_current_version, minor_current_version, patch_current_version = int(split_current_version[0]), int(
        split_current_version[1]), int(split_current_version[2])
    if patch:
        patch_current_version += 1
    if minor:
        patch_current_version = 0
        minor_current_version += 1
    if major:
        patch_current_version = 0
        minor_current_version = 0
        major_current_version += 1

    return '.'.join([str(major_current_version), str(minor_current_version), str(patch_current_version)])


def main(args: List[str]) -> None:
    if len(args) < 2:
        raise CliException

    mode = args[1]
    execute = not (len(args) < 3 or args[2] != '--execute')

    modes = [PATCH, MINOR, MAJOR]
    if mode not in modes:
        raise CliException

    patch = mode == PATCH
    minor = mode == MINOR
    major = mode == MAJOR

    try:
        with open('../version', "r+", encoding='utf-8') as version_f:
            current_version = version_f.readlines()[0]
            print(current_version)

            new_version = generate_version(current_version, patch, minor, major)

            print(f"[-] Update version {current_version} -> {new_version}")

            if execute:
                version_f.seek(0)
                version_f.write(new_version)
                version_f.truncate()
                version_f.flush()
    except FileNotFoundError:
        print('[*] Create new version file')
        with open('../version', "w+", encoding='utf-8') as version_f:
            version_f.write('0.0.1')
    except IOError as e:
        print(f'[*] IOError: {e}')

    if not execute:
        print('[*] no changes were saved under the dry mode. Please add \'--execute\' args to confirm changes.')


if __name__ == '__main__':
    try:
        main(sys.argv)
    except CliException:
        print('[*] Usage: release_version.py [patch|minor|major] [--execute]')
    except Exception as e:
        print(f'[*] Error: {e}')
        sys.exit(1)
