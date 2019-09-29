all:
	$(MAKE) -C cmd/services
	$(MAKE) -C cmd/manager
	$(MAKE) -C cmd/admin

manager:
	$(MAKE) -C cmd/manager

services:
	$(MAKE) -C cmd/services

admin:
	$(MAKE) -C cmd/admin