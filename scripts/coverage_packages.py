import re


def read_lines_from_file(filename):
    """Read lines from a file and return them as list."""
    with open(filename, 'r') as file:
        return [line.strip() for line in file.readlines()]


def filter_packages(packages, patterns):
    """Filter packages to exclude those matching any of the given patterns."""
    included = []
    for package in packages:
        if all(not re.search(pattern, package) for pattern in patterns):
            included.append(package)
    return included


def main():
    # File paths
    packages_file = '/tmp/packages.txt'
    exclude_patterns_file = './scripts/exclude_patterns.txt'

    # Read packages and exclusion patterns
    packages = read_lines_from_file(packages_file)
    exclude_patterns = read_lines_from_file(exclude_patterns_file)

    # Filter packages based on patterns
    included_packages = filter_packages(packages, exclude_patterns)

    # Output the included packages as a comma-separated list
    print(','.join(included_packages))


if __name__ == "__main__":
    main()
