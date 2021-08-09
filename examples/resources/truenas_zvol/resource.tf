resource "truenas_zvol" "zv" {
  pool = "Tank"
  name = "TF ZVOL"
  volsize = 1024 * 1024 * 1024 // 1GiB
  comments = "Test comment"
  compression = "lz4"
}
