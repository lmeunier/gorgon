Gorgon
======

Overview
--------

Gorgon is a `Persona/BrowserId <https://persona.org/>`_ Identity Provider (IdP)
written with the `Go Programming Language <http://golang.org/>`_.

Gorgon is yet in active development state. Do **NOT** use it in production.

Install
-------

You can install Gorgon using a released version (download a tarball from the
`Releases page on Github <https://github.com/lmeunier/gorgon/releases>`_), or
you can `build Gorgon from sources <#build>`_.

- extract the tarball

.. code:: bash

   tar xaf gorgon-0.1.0.tar.gz
   cd gorgon-0.1.0

- create a private and a public keys

.. code:: bash

   openssl genrsa -out private-key.pem 2048
   openssl rsa -in private-key.pem -pubout > public-key.pem

- copy and edit the default configuration file

.. code:: bash

   cp gorgon.ini.example gorgon.ini
   $EDITOR gorgon.ini

Run
---

Once Gorgon is `installed and configured <#install>`_, you are ready to run it.
To start Gorgon, you just have to invoke the ``./gorgon`` command in the folder
where Gorgon is installed.

Gorgon will not daemonize itself. To run Gorgon as a background process, you
must use a tool like `Supervisor <http://supervisord.org/>`_ or `systemd
<http://freedesktop.org/wiki/Software/systemd/>`_.

Once started, Gorgon will listen for HTTP requests on the ``interface:port``
defined in the configuration file. It's up to you to configure your webserver
to redirect requests to Gorgon.

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

- create a tarball

.. code:: bash

   make dist

The ``dist`` target of the Makefile will create an tarball archive in the
``dist/`` folder. You can use this tarball to `install Gorgon <#install>`_.

Credits
-------

Gorgon is maintained by `Laurent Meunier <http://www.deltalima.net/>`_.

Licence
-------

Gorgon is Copyright (c) 2014 Laurent Meunier. It is free software, and may be
redistributed under the terms specified in the LICENSE file (a 3-clause BSD
License).
