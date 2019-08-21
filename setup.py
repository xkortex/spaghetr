import os
from setuptools import find_packages, setup
from spaghetrpc.compile_pb import compile_protobufs

"""Note: there is some rather crazy hackery that goes on here to get the 
following things all into alignment:
- the directory structure
- the way compile_protobufs is fed paths
- the same of the package (including dots)

As such:
- You should typically run from the parent of the pkgname dir (though this will
    affect semantics of `from spaghetrpc.protobuf import time_pb2` since it 
    will preferentially look for a dir named 'spaghetrpc') 

Also, the way compile_protobufs is imported is less than ideal, however it 
allows for `pip install /absolute/path/`, which is one of my criteria for whether
a project structure is well-behaved or not.  

Probably makes sense to eventually put this in a wheel in a pre-stage and then
pull that in to the main dockerfile. 

This should not be needed but may be if you change the pattern:
package_dir={'spaghetrpc.protobuf': 'spaghetrpc/protobuf'},

"""
pkgname = 'spaghetrpc'


def package_files(directories):
    if isinstance(directories, str):
        directories = [directories]
    paths = []
    for directory in directories:
        for (path, directories, filenames) in os.walk(directory):
            for filename in filenames:
                paths.append(os.path.join('..', path, filename))
    return paths

data_files = [
]

package_data = [
]


# Currently using symlinks to make directory structure look more like a package
# since package_dir is not behaving properly with pip -e.
packages = find_packages(exclude=['src', 'src.*'])
print('<Packages>:', packages)

# common dependencies
# todo: fully test unified dependencies
deps = [
    'grpcio>=1.22',
    'vprint'
]

compile_protobufs(os.path.join(pkgname, 'protobuf'))

setup(
    name=pkgname,
    version='0.0.1',
    script_name='setup.py',
    python_requires='>3.5',
    install_requires=deps,
    zip_safe=False,
    packages=[pkgname],
    data_files=data_files,
    include_package_data=True,
    extras_require={
    }
)
