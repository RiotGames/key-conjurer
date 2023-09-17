output https_listener_arn {
	value = aws_lb_listener.https.arn
}

output dns_name {
	value = aws_lb.keyconjurer.dns_name
}
