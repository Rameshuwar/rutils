#!/bin/bash
# Patch for nginx-configurator.sh
sed -i 's/gather_inputs//g' nginx-configurator.sh
sed -i 's/verify_vps//g' nginx-configurator.sh
sed -i 's/verify_dns//g' nginx-configurator.sh
sed -i 's/verify_application//g' nginx-configurator.sh
