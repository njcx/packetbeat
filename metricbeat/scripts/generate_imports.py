import os
import sys

# Generates the file list.go with all modules and metricsets

header = """/*
Package include imports all Module and MetricSet packages so that they register
their factories with the global registry. This package can be imported in the
main package to automatically register all of the standard supported Metricbeat
modules.
*/
package include

import (
\t// This list is automatically generated by `make imports`
"""


def generate(go_beat_path):

    base_dir = "module"
    path = os.path.abspath("module")
    list_file = header

    # Fetch all modules
    for module in sorted(os.listdir(base_dir)):

        if os.path.isfile(path + "/" + module) or module == "_meta":
            continue

        list_file += '	_ "' + go_beat_path + '/module/' + module + '"\n'

        # Fetch all metricsets
        for metricset in sorted(os.listdir(base_dir + "/" + module)):
            if os.path.isfile(base_dir + "/" + module + "/" + metricset) or metricset == "_meta" or metricset == "vendor":
                continue

            list_file += '	_ "' + go_beat_path + '/module/' + module + '/' + metricset + '"\n'

    list_file += ")"

    # output string so it can be concatenated
    print(list_file)

if __name__ == "__main__":
    # First argument is the beat path under GOPATH.
    # (e.g. packetbeat/metricbeat)
    go_beat_path = sys.argv[1]

    generate(go_beat_path)
