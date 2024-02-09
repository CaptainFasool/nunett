package docker

import (
	"context"
	"strings"
	"time"

	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/models"
)

func StartCleanup() {
	DoneCleanup = make(chan bool)
	interval := time.Duration(config.GetConfig().Job.CleanupInterval) * time.Hour * 24
	ticker := time.NewTicker(interval)
	ctx := context.Background()

	go func() {
		defer close(DoneCleanup)
		defer func() {
			if r := recover(); r != nil {
				zlog.Sugar().Errorf("Recovered from error: %v", r)
			}
		}()
		cleanup(ctx)
		for {
			select {
			case <-DoneCleanup:
				return
			case <-ticker.C:
				cleanup(ctx)
			}
		}
	}()
}

func cleanupExitedContainers(ctx context.Context, imageID string) (int, error) {
	containers, err := GetContainersFromImage(ctx, imageID)
	removed := 0
	if err == nil {
		for _, container := range containers {
			if container.State == "exited" {
				if removeErr := StopAndRemoveContainer(ctx, container.ID); removeErr == nil {
					removed++
				}
			}
		}
	} else {
		zlog.Sugar().Errorf("Error fetching containers for image %v: %v", imageID, err)
	}
	return len(containers) - removed, err
}

func cleanupImage(ctx context.Context, imageID string, latest bool) bool {
	containers, err := cleanupExitedContainers(ctx, imageID)
	if err != nil {
		zlog.Sugar().Infof("unable to cleanup containers for image %v: %v", imageID, err)
	} else if latest {
		zlog.Sugar().Infof("skipping latest image %v", imageID)
	} else if containers > 0 {
		zlog.Sugar().Infof("skipping image %v with %v container(s)", imageID, containers)
	} else {
		zlog.Sugar().Infof("removing image: %v", imageID)
		RemoveImage(ctx, imageID)
		return true
	}
	return false
}

func getOldImages() ([]models.ContainerImages, error) {
	var oldImages []models.ContainerImages
	result := db.DB.Where("(image_id, created_at) NOT IN (?)",
		db.DB.Model(&models.ContainerImages{}).Select("image_id, max(created_at)").
			Group("image_name")).
		Find(&oldImages)

	if result.Error != nil {
		return nil, result.Error
	}

	return oldImages, nil
}

func cleanup(ctx context.Context) {
	type LatestImage struct {
		ImageName string
		Latest    time.Time
	}

	oldImages, err := getOldImages()
	if err != nil {
		zlog.Sugar().Warn("Failed to get dangling images")
	}

	for _, image := range oldImages {
		removed := cleanupImage(ctx, image.ImageID, false)
		if removed {
			db.DB.Delete(&image)
		}
	}

	untrackedImages, err := SearchImagesByRefrence(ctx, `registry.gitlab.com/nunet.*`)
	if err == nil && len(untrackedImages) > 0 {
		for _, image := range untrackedImages {
			latest := false
			for _, tag := range image.RepoTags {
				if strings.HasSuffix(tag, ":latest") {
					latest = true
				}
			}
			cleanupImage(ctx, image.ID, latest)
		}
	}
}
