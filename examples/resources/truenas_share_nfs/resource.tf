resource "truenas_share_nfs" "nfs" {
  paths = [
    "<mount path on host to share>",
  ]
  comment = "<optional comment for this share>"
  hosts = [
    "<optional allowed host (hostname or IP address)>",
  ]
  networks = [
    "<optional allowed network in cidr notation>",
  ]
  alldirs = false
  enabled = true
  maproot_user = null
  maproot_group = null
  mapall_user = null
  mapall_group = null
  quiet = false
  ro = false
  security = [
    "<optional security mechanism (possible: krb5[i/p], sys) (needs NFSv4)>",
    "<optional security mechanism with lower priority>",
  ]
}
