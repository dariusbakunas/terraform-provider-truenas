resource "truenas_cronjob" "cj" {
  user = "root"
  command = "ls"
  description = "tf cron job"
  schedule {
    dom = "*"
    dow = "*"
    hour = "*"
    minute = "5"
    month = "*"
  }
}