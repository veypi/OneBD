err = cfg.DB().Where("id = ?", opts.ID).First(data).Error
