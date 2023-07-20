-- initialization of service_resource_requirements table
-- we have to migrate old version table to this new version
-- reference issue: https://gitlab.com/nunet/device-management-service/-/issues/249

CREATE TABLE IF NOT EXISTS "service_resource_requirements" (
    id INTEGER PRIMARY KEY,
    cpu INTEGER DEFAULT 0,
    ram INTEGER DEFAULT 0,
    v_cpu INTEGER DEFAULT 0,
    hdd INTEGER DEFAULT 0,
);

CREATE TABLE "new_service_resource_requirements" (
    id INTEGER PRIMARY KEY,
    tot_cpu_mhz INTEGER DEFAULT 0,
    ram_mb INTEGER DEFAULT 0,
    vcpu_mhz INTEGER DEFAULT 0,
    disk_mb INTEGER DEFAULT 0,
    core_cpu_mhz INTEGER DEFAULT 0,
    cpu_no INTEGER DEFAULT 0
);


INSERT INTO "new_service_resource_requirements" (id, tot_cpu_mhz, ram_mb, vcpu_mhz, disk_mb)
SELECT 
    id,
    IFNULL(cpu, 0),
    IFNULL(ram, 0),
    IFNULL(v_cpu, 0),
    IFNULL(hdd, 0)
FROM "service_resource_requirements";


DROP TABLE IF EXISTS "service_resource_requirements";


ALTER TABLE "new_service_resource_requirements" RENAME TO "service_resource_requirements";
