package gtksrv

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"

	"topotron/lib/place"
)

// PlacesPage is the home screen showing browsable locations.
type PlacesPage struct {
	// gtk widgets
	page    *adw.NavigationPage
	listBox *gtk.ListBox

	// callback
	OnActivated func(place libplace.Place)
}

// newPlacesPage creates a new [PlacesPage] populated with the default places.
func newPlacesPage() *PlacesPage {
	var receiver PlacesPage

	receiver.listBox = gtk.NewListBox()
	receiver.listBox.SetSelectionMode(gtk.SelectionNone)
	receiver.listBox.AddCSSClass("boxed-list")
	receiver.listBox.SetMarginTop(12)
	receiver.listBox.SetMarginBottom(12)
	receiver.listBox.SetMarginStart(12)
	receiver.listBox.SetMarginEnd(12)

	places := libplace.DefaultPlaces()
	for _, p := range places {
		row := newPlaceRow(p)
		receiver.listBox.Append(row)
	}

	receiver.listBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		if nil == receiver.OnActivated {
			return
		}

		index := row.Index()
		if index < 0 || index >= len(places) {
			return
		}

		receiver.OnActivated(places[index])
	})

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)
	scrolled.SetChild(receiver.listBox)

	clamp := adw.NewClamp()
	clamp.SetMaximumSize(600)
	clamp.SetChild(scrolled)

	header := adw.NewHeaderBar()

	toolbar := adw.NewToolbarView()
	toolbar.AddTopBar(header)
	toolbar.SetContent(clamp)

	receiver.page = adw.NewNavigationPage(toolbar, "Topotron")

	return &receiver
}

// newPlaceRow creates an [adw.ActionRow] for a single [libplace.Place].
func newPlaceRow(place libplace.Place) *adw.ActionRow {
	row := adw.NewActionRow()
	row.SetTitle(place.Name)
	row.SetActivatable(true)

	icon := gtk.NewImageFromIconName(place.Icon)
	row.AddPrefix(icon)

	arrow := gtk.NewImageFromIconName("go-next-symbolic")
	row.AddSuffix(arrow)

	return row
}
