resource "truenas_user" "tu" {
  email             = "jane.doe@example.com"
  full_name         = "Jane Doe"
  create_group      = true
  gid               = 42
  uid               = 42
  name              = "jane_doe"
  home_directory    = "/mnt/tank/Users/JaneDoe"
  home_mode         = "0775"
  smb               = false
  sudo              = false
  sudo_nopasswd     = false
  password_disabled = true
  shell             = "/usr/local/bin/bash"
  ssh_public_key    = "ssh-rsa ... jane.doe@example.com"
}