package images

import (
	"fmt"
	"image"
	"image/jpeg"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "github.com/deepteams/webp"
	"github.com/hashicorp/go-hclog"

	"golang.org/x/image/draw"
)

const (
	jpegQuality = 85
)

type Image struct {
	Thumbnail string
	Original  string
	Name      string
	Format    string
}

type Manager struct {
	folderPath      string
	thumbnailsSize  int
	thumbFolderPath string
	images          []*Image
	imageIndex      map[string]int
	log             hclog.Logger
	mu              sync.RWMutex
}

var supportedExtensions = map[string]string{
	".gif":  "gif",
	".jpg":  "jpeg",
	".jpeg": "jpeg",
	".png":  "png",
	".webp": "webp",
}

func New(folderPath string, thumbnailsSize int, logger hclog.Logger) (*Manager, error) {
	if thumbnailsSize <= 0 {
		return nil, fmt.Errorf("thumbnailsSize must be positive, got %d", thumbnailsSize)
	}

	if folderPath == "~" {
		folderPath = os.Getenv("HOME")
	} else if strings.HasPrefix(folderPath, "~/") {
		folderPath = filepath.Join(os.Getenv("HOME"), folderPath[2:])
	}

	absPath, err := filepath.Abs(folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("folder %s does not exist: %w", absPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", absPath)
	}

	thumbFolderPath := filepath.Join(absPath, ".thumbnails")

	if err := os.MkdirAll(thumbFolderPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create thumbnails folder: %w", err)
	}

	mgr := &Manager{
		folderPath:      absPath,
		thumbnailsSize:  thumbnailsSize,
		thumbFolderPath: thumbFolderPath,
		imageIndex:      make(map[string]int),
	}

	needRegenerate, err := mgr.checkSizeFile()
	if err != nil {
		return nil, err
	}

	if err := mgr.sync(needRegenerate); err != nil {
		return nil, err
	}

	return mgr, nil
}

func (m *Manager) checkSizeFile() (needRegenerate bool, err error) {
	sizePath := filepath.Join(m.thumbFolderPath, ".size")

	data, err := os.ReadFile(sizePath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, m.writeSizeFile()
		}
		return false, err
	}

	var savedSize int
	fmt.Sscanf(string(data), "%d", &savedSize)

	if savedSize != m.thumbnailsSize {
		m.log.Debug(
			"Size changed",
			"Old", savedSize,
			"New", m.thumbnailsSize,
		)
		return true, m.writeSizeFile()
	}

	return false, nil
}

func (m *Manager) writeSizeFile() error {
	sizePath := filepath.Join(m.thumbFolderPath, ".size")
	return os.WriteFile(sizePath, fmt.Appendf(nil, "%d", m.thumbnailsSize), 0644)
}

func (m *Manager) sync(forceRegenerate bool) error {
	originals, err := m.scanOriginals()
	if err != nil {
		return fmt.Errorf("failed to scan originals: %w", err)
	}

	existingThumbs, err := m.scanThumbnails()
	if err != nil {
		return fmt.Errorf("failed to scan thumbnails: %w", err)
	}

	tasks := make([]*Image, 0)
	validOriginals := make([]*Image, 0)
	seenNames := make(map[string]bool)

	for _, img := range originals {
		if seenNames[img.Name] {
			m.log.Debug(
				"duplicate detected, skipping second file",
				"Name", img.Name,
				"File", img.Original,
			)
			continue
		}
		seenNames[img.Name] = true

		thumbPath := filepath.Join(m.thumbFolderPath, img.Name+".jpg")
		img.Thumbnail = thumbPath

		_, thumbExists := existingThumbs[img.Name]

		if forceRegenerate || !thumbExists {
			tasks = append(tasks, img)
		}

		validOriginals = append(validOriginals, img)
		delete(existingThumbs, img.Name)
	}

	for thumbName := range existingThumbs {
		thumbPath := filepath.Join(m.thumbFolderPath, thumbName+".jpg")
		if err := os.Remove(thumbPath); err != nil {
			m.log.Error(
				"Failed to remove orphan thumbnail",
				"File", thumbPath,
				"Error", err,
			)
		} else {
			m.log.Trace(
				"Removed orphan thumbnail",
				"File", thumbName,
			)
		}
	}

	if len(tasks) > 0 {
		m.log.Debug(
			"Generating thumbnails",
			"Files", len(tasks),
			"Workers", runtime.NumCPU(),
		)
		if err := m.generateThumbnailsParallel(tasks); err != nil {
			return err
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.images = validOriginals
	m.imageIndex = make(map[string]int)
	for i, img := range m.images {
		m.imageIndex[img.Name] = i
	}

	m.log.Trace(
		"Sync completed",
		"Total", len(m.images),
	)
	return nil
}

func (m *Manager) scanOriginals() ([]*Image, error) {
	entries, err := os.ReadDir(m.folderPath)
	if err != nil {
		return nil, err
	}

	var images []*Image
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		ext := strings.ToLower(filepath.Ext(filename))
		format, ok := supportedExtensions[ext]
		if !ok {
			continue
		}

		name := filename[:len(filename)-len(ext)]
		originalPath := filepath.Join(m.folderPath, filename)

		images = append(images, &Image{
			Name:     name,
			Original: originalPath,
			Format:   format,
		})
	}

	return images, nil
}

func (m *Manager) scanThumbnails() (map[string]bool, error) {
	entries, err := os.ReadDir(m.thumbFolderPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]bool), nil
		}
		return nil, err
	}

	thumbnails := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".jpg") {
			name := entry.Name()[:len(entry.Name())-4]
			thumbnails[name] = true
		}
	}

	return thumbnails, nil
}

func (m *Manager) generateThumbnailsParallel(images []*Image) error {
	numWorkers := runtime.NumCPU()
	if numWorkers < 1 {
		numWorkers = 1
	}

	tasks := make(chan *Image, len(images))
	errCh := make(chan error, len(images))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for img := range tasks {
				thumbPath := filepath.Join(m.thumbFolderPath, img.Name+".jpg")
				if err := createThumbnail(img.Original, thumbPath, m.thumbnailsSize); err != nil {
					errCh <- fmt.Errorf("worker %d: failed to create thumbnail for '%s': %w", workerID, img.Name, err)
				}
			}
		}(i)
	}

	for _, img := range images {
		tasks <- img
	}
	close(tasks)

	wg.Wait()
	close(errCh)

	var hasErrors bool
	for err := range errCh {
		m.log.Error("Generate thumbnail", "Error", err)
		hasErrors = true
	}

	if hasErrors {
		return fmt.Errorf("some thumbnails failed to generate (see logs)")
	}
	return nil
}

func (m *Manager) GetList() []*Image {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Image, len(m.images))
	copy(result, m.images)
	return result
}

func (m *Manager) Get(name string) *Image {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if idx, ok := m.imageIndex[name]; ok {
		return m.images[idx]
	}
	return nil
}

func (m *Manager) Rescan() error {
	m.log.Debug("Rescanning folder...")
	return m.sync(false)
}

func (m *Manager) Random() *Image {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := len(m.images)
	if count == 0 {
		m.log.Warn("Random() called with no images available")
		return nil
	}

	return m.images[rand.Intn(count)]
}

func createThumbnail(originalPath, thumbPath string, targetHeight int) error {
	file, err := os.Open(originalPath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	if origWidth == 0 || origHeight == 0 {
		return fmt.Errorf("invalid image dimensions: %dx%d", origWidth, origHeight)
	}

	var newWidth, newHeight int
	if origHeight <= targetHeight {
		newWidth = origWidth
		newHeight = origHeight
	} else {
		newHeight = targetHeight
		newWidth = int(float64(origWidth) * float64(targetHeight) / float64(origHeight))
	}

	if newWidth <= 0 {
		newWidth = 1
	}
	if newHeight <= 0 {
		newHeight = 1
	}

	dstRGBA := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	draw.ApproxBiLinear.Scale(dstRGBA, dstRGBA.Bounds(), img, bounds, draw.Over, nil)

	if err := os.MkdirAll(filepath.Dir(thumbPath), 0755); err != nil {
		return err
	}

	out, err := os.Create(thumbPath)
	if err != nil {
		return err
	}
	defer out.Close()

	return jpeg.Encode(out, dstRGBA, &jpeg.Options{Quality: jpegQuality})
}
