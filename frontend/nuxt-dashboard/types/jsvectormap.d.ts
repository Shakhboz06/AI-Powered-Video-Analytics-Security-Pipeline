declare module 'jsvectormap' {
  interface MapOptions {
    selector: string | HTMLElement
    map: string
    zoomButtons?: boolean
    regionStyle?: {
      initial?: Record<string, any>
      hover?: Record<string, any>
      selected?: Record<string, any>
      selectedHover?: Record<string, any>
    }
    markers?: Array<{
      name: string
      coords: [number, number]
    }>
    markerStyle?: {
      initial?: Record<string, any>
      hover?: Record<string, any>
      selected?: Record<string, any>
      selectedHover?: Record<string, any>
    }
    onRegionTooltipShow?: (event: MouseEvent, tooltip: any, code: string) => void
    [key: string]: any
  }

  class jsVectorMap {
    constructor(options: MapOptions)
    destroy(): void
    reset(): void
    setFocus(config: any): void
    updateSize(): void
    addMarkers(markers: any[]): void
    removeMarkers(markers: any[]): void
    [key: string]: any
  }

  export default jsVectorMap
}

declare module 'jsvectormap/dist/maps/world' {
  const world: any
  export default world
}

declare module 'jsvectormap/dist/css/jsvectormap.css' {}