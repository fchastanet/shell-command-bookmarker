#!/usr/bin/env bash

awk \
  '''
    1;/---/{
      print "###############################################################################"
      print "# AUTOMATICALLY GENERATED"
      print "# DO NOT EDIT IT"
      print "# @generated"
      print "###############################################################################"
    }
  ''' .pre-commit-config.yaml >.pre-commit-config-github.yaml
sed -i -E \
  -e '0,/fail_fast: true/s//fail_fast: false/' \
  -e 's/stages: \[\] # GITHUB/stages: \[manual\] # GITHUB/' \
  .pre-commit-config-github.yaml
