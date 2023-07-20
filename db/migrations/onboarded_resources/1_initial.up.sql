-- initialization of onboarded_resources table
-- we have to migrate old version table to this new version
-- (also changing name from available_resources to onboarded_resources)
-- reference issue: https://gitlab.com/nunet/device-management-service/-/issues/249

CREATE TABLE IF NOT EXISTS "available_resources" (
    id INTEGER PRIMARY KEY,
    tot_cpu_hz INTEGER DEFAULT 0,
    ram INTEGER DEFAULT 0,
    vcpu INTEGER DEFAULT 0,
    disk INTEGER DEFAULT 0,
    cpu_no INTEGER DEFAULT 0,
    cpu_hz REAL DEFAULT 0.0,
    price_cpu INTEGER DEFAULT 0,
    price_ram INTEGER DEFAULT 0,
    price_disk INTEGER DEFAULT 0
);


CREATE TABLE "onboarded_resources" (
    id INTEGER PRIMARY KEY,
    tot_cpu_mhz INTEGER DEFAULT 0,
    ram_mb INTEGER DEFAULT 0,
    vcpu_mhz INTEGER DEFAULT 0,
    disk_mb INTEGER DEFAULT 0,
    core_cpu_mhz INTEGER DEFAULT 0,
    cpu_no INTEGER DEFAULT 0
);


INSERT INTO "onboarded_resources" (id, tot_cpu_mhz, ram_mb, vcpu_mhz, disk_mb, core_cpu_mhz, cpu_no)
SELECT 
    id,
    IFNULL(tot_cpu_hz, 0),
    IFNULL(ram, 0),
    IFNULL(vcpu, 0),
    IFNULL(disk, 0),
    IFNULL(cpu_hz, 0),
    IFNULL(cpu_no, 0)
FROM "available_resources";


DROP TABLE IF EXISTS "available_resources";
