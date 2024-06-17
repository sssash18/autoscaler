### Creating Nodes in virtual cluster

#### Option 1
- this CA binary will be passed an additional flag `gardener-shoot-name`.
- It can connect to the shoot and get the shoot yaml
- It knows about nodegroups, workerpools, etc.
- This info can be used for creating node objects during scaleups in virtual cluster.

#### Option 2
- We have a small script - `grab_cluster_info`.
- This will connect to gardener cluster and download relevant yaml files (shoot yaml,ca yaml,mcd yaml,nodes yaml).
- CA can be configured with path `gardener-cluster-info=<path_to_folder>`.
- local provider can now launch virtual nodes using this info.

#### Option 3
- CA can be configured with a custom param `gardener-cluster-hist=<path_to_db>` .