# controller-metrics

Standardized metrics for kubernetes/OpenShift controllers.

- [controller-metrics](#controller-metrics)
  - [Overview](#overview)
  - [Packages](#packages)

## Overview

This is a go module providing packages that help kubernetes and/or OpenShift controllers expose
generally-useful Prometheus metrics in a standardized way. The idea is twofold:

- As with any module, this saves you the time and effort of developing the inner workings of these 
  pieces.
- If your organization maintains more than one operator, you can use this module across all of them
  to provide a consistent, predictable experience to their consumers.

## Packages

- [pkg/apicall](pkg/apicall/README.md) times and counts calls to external APIs.
