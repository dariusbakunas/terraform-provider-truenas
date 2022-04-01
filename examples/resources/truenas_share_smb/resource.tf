resource "truenas_share_smb" "smb" {
  path = "<mount path on host to share>"
  path_suffix = "<optional suffix to append to mount path>"
  purpose = "<optional TrueNAS preset ID (might lock some settings!)>"
  comment = "<optional comment for this share>"
  hostsallow = [
    "<optional CIDR host/network to allow access for>",
  ]
  hostsdeny = [
    "<optional CIDR host/network to deny access for or 'ALL' for whitelist model>",
  ]
  name = "<optional share name, defaults to last path segment>"
  ro = true
  aapl_name_mangling = false
  abe = false
  acl = true
  auxsmbconf = ""
  browsable = false
  durablehandle = false
  enabled = true
  fsrvp = false
  guestok = false
  home = false
  recyclebin = false
  shadowcopy = true
  streams = false
  timemachine = false
}
