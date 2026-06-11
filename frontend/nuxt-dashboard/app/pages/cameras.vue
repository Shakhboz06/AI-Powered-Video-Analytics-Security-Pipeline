<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Configuration"
      title="Cameras"
      subtitle="Manage configured camera sources used by monitoring, zones, alerts, and analytics."
    >
      <template #actions>
        <button
          type="button"
          class="btn-primary"
          @click="openCreate"
        >
          <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
          Add camera
        </button>
      </template>
    </PageHeader>

    <div v-if="pageError" class="app-banner-error">{{ pageError }}</div>

    <section class="app-card overflow-hidden">
      <div class="border-b border-white/[0.06] px-4 py-3 flex items-center justify-between gap-3">
        <h2 class="app-section-title">Camera list</h2>
        <div class="flex items-center gap-3">
          <button
            type="button"
            class="text-xs text-teal-300 hover:underline disabled:text-gray-600"
            :disabled="loading"
            @click="loadCameras"
          >
            {{ loading ? 'Refreshing…' : 'Refresh' }}
          </button>
          <button
            type="button"
            class="btn-primary text-sm"
            @click="openCreate"
          >
            <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
            Add
          </button>
        </div>
      </div>

      <div v-if="loading && !cameras.length" class="divide-y divide-white/[0.06]">
        <div v-for="i in 4" :key="i" class="flex items-center gap-4 px-4 py-4">
          <div class="flex-1 space-y-2">
            <Skeleton width="8rem" height="0.9rem" />
            <Skeleton width="4rem" height="0.7rem" />
          </div>
          <Skeleton width="14rem" height="0.8rem" />
          <Skeleton width="4rem" height="1.5rem" rounded="rounded-md" />
          <Skeleton width="6rem" height="0.8rem" />
        </div>
      </div>
      <EmptyState
        v-else-if="!cameras.length"
        icon="camera"
        title="No cameras configured yet"
        message="Add your first camera source to start monitoring detections, zones, and alerts."
      >
        <template #action>
          <button type="button" class="btn-primary" @click="openCreate">
            <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
            Add camera
          </button>
        </template>
      </EmptyState>
      <div v-else class="app-table-wrap">
        <table class="app-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Video source</th>
              <th>Status</th>
              <th>Updated</th>
              <th class="text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="camera in cameras" :key="camera.camera_id" class="group">
              <td>
                <div class="flex items-center gap-3">
                  <span
                    class="grid h-9 w-9 shrink-0 place-items-center rounded-xl border transition-colors"
                    :class="camera.is_active ? 'border-teal-500/20 bg-teal-950/20' : 'border-white/[0.06] bg-white/[0.02]'"
                  >
                    <svg class="h-4 w-4" :class="camera.is_active ? 'text-teal-400' : 'text-gray-500'" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>
                  </span>
                  <div>
                    <p class="font-medium text-gray-100">{{ camera.camera_name }}</p>
                    <p class="mt-0.5 text-xs text-gray-500 font-mono">ID {{ camera.camera_id }}</p>
                  </div>
                </div>
              </td>
              <td class="max-w-md">
                <p class="truncate font-mono text-xs text-gray-300" :title="camera.video_source">
                  {{ camera.video_source }}
                </p>
              </td>
              <td>
                <span
                  class="stat-pill"
                  :class="camera.is_active ? 'border-emerald-800/50 bg-emerald-950/30 text-emerald-200' : 'text-gray-400'"
                >
                  <span class="h-1.5 w-1.5 rounded-full" :class="camera.is_active ? 'bg-emerald-400 shadow-[0_0_6px_rgba(52,211,153,0.5)]' : 'bg-gray-600'" />
                  {{ camera.is_active ? 'Active' : 'Inactive' }}
                </span>
              </td>
              <td class="text-xs text-gray-500">
                {{ formatDate(camera.updated_at || camera.created_at) }}
              </td>
              <td>
                <div class="flex justify-end gap-2">
                  <button type="button" class="btn-ghost-sm" @click="openEdit(camera)">
                    <svg class="mr-1 inline h-3 w-3" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125" /></svg>
                    Edit
                  </button>
                  <button type="button" class="btn-danger-sm" @click="onDelete(camera)">
                    <svg class="mr-1 inline h-3 w-3" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" /></svg>
                    Delete
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <AppModal
      :open="showModal"
      :title="modalMode === 'create' ? 'Add camera' : 'Edit camera'"
      subtitle="Configure a video source for detection and alerting."
      @close="closeModal"
    >
      <div class="space-y-4">
        <div>
          <label class="app-label mb-1.5 block">Camera name <span class="text-red-400">*</span></label>
          <input v-model="form.camera_name" type="text" class="app-input w-full" placeholder="cam1">
        </div>
        <div>
          <label class="app-label mb-1.5 block">Video source <span class="text-red-400">*</span></label>
          <input v-model="form.video_source" type="text" class="app-input w-full" placeholder="rtsp://camera.local/stream">
        </div>
        <label class="inline-flex items-center gap-2 text-sm text-gray-300">
          <input v-model="form.is_active" type="checkbox" class="app-checkbox">
          Active
        </label>
      </div>
      <template #footer>
        <button type="button" class="btn-ghost" :disabled="saving" @click="closeModal">Cancel</button>
        <button
          type="button"
          class="btn-primary disabled:opacity-50"
          :disabled="saving || !form.camera_name.trim() || !form.video_source.trim()"
          @click="submit"
        >
          {{ saving ? 'Saving…' : 'Save' }}
        </button>
      </template>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import PageHeader from '~/components/ui/PageHeader.vue'
import AppModal from '~/components/ui/AppModal.vue'
import Skeleton from '~/components/ui/Skeleton.vue'
import EmptyState from '~/components/ui/EmptyState.vue'
import type { CameraConfig } from '~/types/security'

definePageMeta({
  layout: 'security',
  middleware: 'auth',
})

const api = useApi()
const toast = useToast()
const confirm = useConfirm()

const cameras = ref<CameraConfig[]>([])
const loading = ref(false)
const saving = ref(false)
const pageError = ref<string | null>(null)

const showModal = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingId = ref<number | null>(null)

const form = reactive({
  camera_name: '',
  video_source: '',
  is_active: true,
})

onMounted(() => {
  void loadCameras()
})

function resetForm() {
  form.camera_name = ''
  form.video_source = ''
  form.is_active = true
  editingId.value = null
}

function openCreate() {
  resetForm()
  modalMode.value = 'create'
  showModal.value = true
}

function openEdit(camera: CameraConfig) {
  modalMode.value = 'edit'
  editingId.value = camera.camera_id
  form.camera_name = camera.camera_name
  form.video_source = camera.video_source
  form.is_active = camera.is_active
  showModal.value = true
}

function closeModal() {
  if (saving.value) return
  showModal.value = false
}

function formatDate(iso?: string) {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })
}

async function loadCameras() {
  loading.value = true
  pageError.value = null
  try {
    const res = await api.getCameraConfigs()
    cameras.value = res.cameras ?? []
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Failed to load cameras'
    cameras.value = []
  }
  finally {
    loading.value = false
  }
}

async function submit() {
  saving.value = true
  pageError.value = null
  const body = {
    camera_name: form.camera_name.trim(),
    video_source: form.video_source.trim(),
    is_active: form.is_active,
  }
  try {
    if (modalMode.value === 'create') {
      await api.createCamera(body)
    }
    else if (editingId.value != null) {
      await api.updateCamera(editingId.value, body)
    }
    showModal.value = false
    await loadCameras()
    toast.success(modalMode.value === 'create' ? 'Camera added' : 'Camera updated', { description: body.camera_name })
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Failed to save camera'
    toast.error('Failed to save camera')
  }
  finally {
    saving.value = false
  }
}

async function onDelete(camera: CameraConfig) {
  const ok = await confirm.ask({
    title: 'Delete camera?',
    message: `"${camera.camera_name}" will be permanently removed along with its association to zones and alerts.`,
    confirmLabel: 'Delete',
    danger: true,
  })
  if (!ok) return
  pageError.value = null
  try {
    await api.deleteCamera(camera.camera_id)
    await loadCameras()
    toast.success('Camera deleted', { description: camera.camera_name })
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Failed to delete camera'
    toast.error('Failed to delete camera')
  }
}
</script>
