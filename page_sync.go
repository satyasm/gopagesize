package main

func (p *page) resolve() error {
	body, err := p.base.get()
	if err != nil {
		return err
	}
	p.total += p.base.size
	if err := p.parseResources(body); err != nil {
		return err
	}
	return p.resolveResources()
}

func (p *page) resolveResources() error {
	for _, a := range p.assets {
		for _, r := range a {
			if _, err := r.get(); err == nil {
				p.total += r.size
			}
		}
	}
	return nil
}
