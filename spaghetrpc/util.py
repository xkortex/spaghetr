#!/usr/bin/env python
# -*- coding: utf-8 -*-


def parse_nullable(nullable_msg):
    if nullable_msg.ListFields():
        return nullable_msg.val
    return None
