Gorgon
======

Overview
--------

Gorgon is a `Persona/BrowserId <https://persona.org/>`_ Identity Provider (IdP)
written with the `Go Programming Language <http://golang.org/>`_.

Gorgon is yet in active development state. Do **NOT** use it in production.

Build
-----

- initialize a workspace directory and set ``GOPATH`` and ``PATH`` accordingly

.. code:: bash

    mkdir -p "$HOME/gorgon/gopath"
    export GOPATH="$HOME/gorgon/gopath"
    export PATH="$GOPATH/bin:$PATH"

- install Gorgon sources

.. code:: bash

    go get -d github.com/lmeunier/gorgon

- build Gorgon

.. code:: bash

    cd "$GOPATH/src/github.com/lmeunier/gorgon"
    make install_deps
    make build

The ``build`` target of the Makefile will create an ``gorgon`` executable file
in the current folder.

Credits
-------

Gorgon is maintained by `Laurent Meunier <http://www.deltalima.net/>`_.

Licence
-------

Gorgon is Copyright (c) 2014 Laurent Meunier. It is free software, and may be
redistributed under the terms specified in the LICENSE file (a 3-clause BSD
License).
