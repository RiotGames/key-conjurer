resource "aws_waf_ipset" "ipset" {
  count = var.create_waf_acl == true ? 1 : 0

  name = "keyconjurer-${terraform.workspace}-tfIPSet"
  ip_set_descriptors {
    type  = "IPV4"
    value = "127.0.0.1/32"
  }
}

resource "aws_waf_rule" "not_ip_rule" {
  count = var.create_waf_acl == true ? 1 : 0

  name        = "KeyConjurer${terraform.workspace}WafRule"
  metric_name = "KeyConjurer${terraform.workspace}WafRule"

  predicates {
    data_id = aws_waf_ipset.ipset[0].id
    negated = true
    type    = "IPMatch"
  }
}

resource "aws_waf_web_acl" "keyconjurer_waf_acl" {
  count = var.create_waf_acl == true ? 1 : 0

  name        = "KeyConjurerWAF${terraform.workspace}WebACL"
  metric_name = "KeyConjurerWAF${terraform.workspace}WebACL"

  default_action {
    type = "ALLOW"
  }

  rules {
    action {
      type = "BLOCK"
    }

    priority = 1
    rule_id  = aws_waf_rule.not_ip_rule[0].id
    type     = "REGULAR"
  }
}

