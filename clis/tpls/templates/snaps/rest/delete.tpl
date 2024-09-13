err = cfg.DB().Where("id = ?", opts.ID).Delete(data).Error
