package widgets

type Site struct {
	main_page *MainPage
}

func NewSite() *Site {
	self := &Site{}
	self.main_page = NewMainPage(self)
	return self
}

func (self *Site) ApplyToDocument() error {
	self.main_page.RenderMainPage()
	return nil
}
