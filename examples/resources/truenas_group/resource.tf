resource "truenas_group" "tg" {
  allow_duplicate_gid = false
  gid                 = 42
  name                = "ExampleGroup"
  smb                 = false
  sudo                = false
  sudo_nopasswd       = false
  sudo_commands       = []
}