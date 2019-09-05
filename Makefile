all:
	$(MAKE) -C cmd/services

manager:
	$(MAKE) -C cmd/manager

services:
	$(MAKE) -C cmd/services

admin:
	$(MAKE) -C cmd/admin