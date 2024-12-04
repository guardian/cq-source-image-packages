# Table: amigo_bake_packages

This table shows data for Amigo Bake Packages.

The primary key for this table is **_cq_id**.

## Columns

| Name          | Type          |
| ------------- | ------------- |
|_cq_id (PK)|`uuid`|
|_cq_parent_id|`uuid`|
|base_name|`utf8`|
|base_ami_id|`utf8`|
|base_eol_date|`timestamp[us, tz=UTC]`|
|recipe_id|`utf8`|
|bake_number|`int64`|
|source_ami_id|`utf8`|
|started_at|`timestamp[us, tz=UTC]`|
|started_by|`utf8`|
|package_name|`utf8`|
|package_version|`utf8`|