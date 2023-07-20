-- initialization of free_resources table
-- we have to migrate old version table to this new version
-- reference issue: https://gitlab.com/nunet/device-management-service/-/issues/249

CREATE TABLE IF NOT EXISTS "free_resources" (
    id INTEGER PRIMARY KEY,
    tot_cpu_hz INTEGER DEFAULT 0,
    ram INTEGER DEFAULT 0,
    vcpu INTEGER DEFAULT 0,
    disk INTEGER DEFAULT 0,
    price_cpu INTEGER DEFAULT 0,
    price_ram INTEGER DEFAULT 0,
    price_disk INTEGER DEFAULT 0
);


CREATE TABLE "free_resources_new" (
    id INTEGER PRIMARY KEY,
    tot_cpu_mhz INTEGER DEFAULT 0,
    ram_mb INTEGER DEFAULT 0,
    vcpu_mhz INTEGER DEFAULT 0,
    disk_mb INTEGER DEFAULT 0,
    core_cpu_mhz INTEGER DEFAULT 0,
    cpu_no INTEGER DEFAULT 0
);


INSERT INTO "free_resources_new" (id, tot_cpu_mhz, ram_mb, vcpu_mhz, disk_mb)
SELECT 
    id,
    IFNULL(tot_cpu_hz, 0),
    IFNULL(ram, 0),
    IFNULL(vcpu, 0),
    IFNULL(disk, 0)
FROM "free_resources";


DROP TABLE IF EXISTS "free_resources";


ALTER TABLE "free_resources_new" RENAME TO "free_resources";
