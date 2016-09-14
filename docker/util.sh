#!/bin/bash

#Get the petset name based on our current hostname
getPetSetName(){
  HOSTNAME=$(hostname -s)
  GROUPNAME=$(echo $INPUT| cut -d '-' -f 1)
  return GROUPNAME
}
