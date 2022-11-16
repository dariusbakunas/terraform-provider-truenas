resource "truenas_zvol" "zv" {
  pool = "Tank"
  name = "TestZVOL"
  volsize = 1024 * 1024 * 1024 // 1GiB
  comments = "Test comment"
  compression = "lz4"
}
