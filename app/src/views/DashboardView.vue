<script lang="ts" setup>
import ArrowTopIcon from '@/components/icons/ArrowTopIcon.vue';
import CheckIcon from '@/components/icons/CheckIcon.vue';
import LinkIcon from '@/components/icons/LinkIcon.vue';
import LoadingIcon from '@/components/icons/LoadingIcon.vue';
import TimesIcon from '@/components/icons/TimesIcon.vue';
import { DashboardTransport } from '@/transport/dashboard';
import { computed, onBeforeMount, ref, watch } from 'vue';
declare const Telegram: any


enum TaskStatus {
  Forming = 'Формируется',
  CanDo = 'Можно делать',
  OnHold = 'На паузе',
  Waiting = 'Ожидание',
  InProgress = 'В работе',
  NeedDiscussion = 'Надо обсудить',
  CodeReview = 'Код-ревью',
  InternalCheck = 'Внутренняя проверка',
  ReadyToUpload = 'Можно выгружать',
  ClientCheck = 'Проверка клиентом',
  Cancelled = 'Отменена',
  Done = 'Готова'
}

enum TaskStatusCategory {
  ToDo = 'Необходимо делать',
  InProgress = 'В работе',
  Done = 'Готово'
}

const getCategoryByStatus = (status: TaskStatus): TaskStatusCategory => {
  switch (status) {
    case TaskStatus.Forming:
    case TaskStatus.CanDo:
    case TaskStatus.OnHold:
    case TaskStatus.Waiting:
      return TaskStatusCategory.ToDo;
    case TaskStatus.InProgress:
    case TaskStatus.NeedDiscussion:
    case TaskStatus.CodeReview:
      return TaskStatusCategory.InProgress;
    case TaskStatus.InternalCheck:
    case TaskStatus.ReadyToUpload:
    case TaskStatus.ClientCheck:
    case TaskStatus.Cancelled:
    case TaskStatus.Done:
      return TaskStatusCategory.Done;
    default:
      throw new Error('Unknown task status');
  }
}

interface Task {
  title: string;
  status: TaskStatus;
}

const currentQuarter = Math.ceil((new Date().getMonth() + 1) / 3)

const querterTasks = ref<Task[]>([])
const quarterTasksDone = computed(() => querterTasks.value.filter(task => getCategoryByStatus(task.status) === TaskStatusCategory.Done).length)


onBeforeMount(async () => {
  countUserProgress()
  querterTasks.value = await DashboardTransport.getQuarterTasks()

  // const tg = Telegram.WebApp

  // tg.expand()
  // tg.isVerticalSwipesEnabled = false
  // tg.disableVerticalSwipes()
  // tg.ready()

  projectsWithSheets.value = await DashboardTransport.listProjectsWithSheets()
})

const fileInput = ref<HTMLInputElement | null>(null);
const triggerFileUpload = () => {
  fileInput.value?.click();
};

const handleFileUpload = (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (target.files && target.files.length > 0) {
    const file = target.files[0];
    const reader = new FileReader();
    reader.onload = (e) => {
      mindmapLoaderRequestProducer.start(async () => {
        try {
          await DashboardTransport.parseMindmap(file)
          Message.newMessage(MessageType.Success, "Майндкарта загружена")
        } catch (error) {
          console.error(error)
          Message.newMessage(MessageType.Error, "Не удалось загрузить майндкарту")

        }
      })
    };
    reader.readAsText(file);
  }
}

enum Period {
  Quarter = 'Квартал',
  Month = 'Месяц',
  Week = 'Неделя'
}

const currentPeriod = ref<Period>(Period.Quarter)

const userTasks = ref<Task[]>([])

const userPercents = ref(0)

const countUserProgress = async () => {
  let periodStart: Date;
  let periodEnd: Date;

  const now = new Date();
  switch (currentPeriod.value) {
    case Period.Quarter: {
      const quarterStartMonth = (currentQuarter - 1) * 3;
      periodStart = new Date(now.getFullYear(), quarterStartMonth, 1, 0, 0, 0);
      periodEnd = new Date(now.getFullYear(), quarterStartMonth + 3, 0, 23, 59, 59);
      break;
    }
    case Period.Month:
      periodStart = new Date(now.getFullYear(), now.getMonth(), 1, 0, 0, 0);
      periodEnd = new Date(now.getFullYear(), now.getMonth() + 1, 0, 23, 59, 59);
      break;
    case Period.Week: {
      const day = now.getDay();
      const diff = now.getDate() - day + (day === 0 ? -6 : 1);
      periodStart = new Date(now);
      periodStart.setDate(diff);
      periodStart.setHours(0, 0, 0, 0);

      periodEnd = new Date(periodStart);
      periodEnd.setDate(periodEnd.getDate() + 6);
      periodEnd.setHours(23, 59, 59, 999);
      break;
    }
  }


  let tasks;
  try {
    tasks = await DashboardTransport.getTasksOfEmployee(Telegram.WebApp.initDataUnsafe.user.username, Math.floor(periodStart.getTime() / 1000), Math.floor(periodEnd.getTime() / 1000))
  } catch (error) {
    console.error(error)
  }
  userTasks.value = tasks

  if (userTasks.value.length === 0) {
    userPercents.value = 0
    return
  }


  const totalDaysInPeriod = (periodEnd.getTime() - periodStart.getTime()) / (1000 * 60 * 60 * 24);
  const daysGoneFromPeriodStart = (now.getTime() - periodStart.getTime()) / (1000 * 60 * 60 * 24);

  const percentOfDaysGone = (daysGoneFromPeriodStart / totalDaysInPeriod) * 100;

  const percentOfTasksDone = (userTasks.value.filter(task => getCategoryByStatus(task.status) === TaskStatusCategory.Done).length / userTasks.value.length) * 100 || 0;


  userPercents.value = percentOfTasksDone - percentOfDaysGone;

}

watch(currentPeriod, countUserProgress)

class RequestProducer {
  state = ref({
    loading: false
  });

  start = async (f: Function) => {
    this.state.value.loading = true;

    try {
      await f();
    } finally {
      this.state.value.loading = false;
    }
  }
}

const toSheetsRequestProducer = new RequestProducer()

const loadToSheets = async () => {
  try {
    await toSheetsRequestProducer.start(DashboardTransport.updateGoogleSheets);
    Message.newMessage(MessageType.Success, "Данные обновлены")
  } catch (error) {
    console.error(error)
    Message.newMessage(MessageType.Error, "Не удалось обновить данные")
  }
}

const mindmapLoaderRequestProducer = new RequestProducer()

const enum MessageType {
  Error = "error",
  Warning = "warning",
  Success = "success",
  Info = "info"
}

class Message {
  type: MessageType;
  text: string;
  id: number;
  static duration = 3000;

  static messages = ref<Map<number, Message>>(new Map());
  static nextId = 0;

  constructor(type: MessageType, text: string) {
    this.type = type;
    this.text = text;
    this.id = Message.nextId++;
  }

  static newMessage(type: MessageType, text: string) {
    const message = new Message(type, text);
    this.messages.value.set(message.id, message);

    setTimeout(() => {
      this.removeMessage(message.id);
    }, this.duration);
  }

  static removeMessage(id: number) {
    this.messages.value.delete(id);
  }
}

const progressRequestProducer = new RequestProducer()

const enum ProgressStatus {
  Good = 'good',
  Bad = 'bad',
  Neutral = 'neutral'
}
const progressToStatus = (num: number) => {
  num = Math.round(num)
  if (num < 0) return ProgressStatus.Bad
  if (num > 0) return ProgressStatus.Good
  return ProgressStatus.Neutral
}

const salaryNotifyRequestProducer = new RequestProducer()

const salaryNotify = async () => {
  try {
    await salaryNotifyRequestProducer.start(DashboardTransport.notifyAboutSalary);
    Message.newMessage(MessageType.Success, "Уведомления отправлены")
  } catch (error) {
    console.error(error)
    Message.newMessage(MessageType.Error, "Не удалось отправить обновления")
  }
}

const sendFeedback = () => {
  Telegram.WebApp.openTelegramLink('https://t.me/incetro')
}

interface project {
  id: string;
  name: string;
  sheetsLink: string;
  icon: string;
  iconType: string;
}

const projectsWithSheets = ref<project[]>([])

const projectSheetsUpdateRequestProducer = new RequestProducer()
const updateProjectSheets = (projectID: string) =>{
  projectSheetsUpdateRequestProducer.start(async () => {
    try {
      await DashboardTransport.updateProjectSheets(projectID);
      Message.newMessage(MessageType.Success, "Данные обновлены");
    } catch (error) {
      console.error(error);
      Message.newMessage(MessageType.Error, "Не удалось обновить данные");
    }
  });

}

</script>

<template>
  <div class="messages">
    <TransitionGroup name="fade-slide">
      <div class="message" v-for="(message, i) of Message.messages.value.values()" :key="`message${i}`"
        @click="Message.removeMessage(message.id)" :class="message.type">
        <div class="msg-icon">
          <TimesIcon v-if="message.type === MessageType.Error" class=" text-danger" />
          <CheckIcon v-else-if="message.type === MessageType.Success" class=" text-success" />
        </div>
        <p>{{ message.text }}</p>
      </div>
    </TransitionGroup>
  </div>
  <main class="">
    <div class="header">
      <h1>Incetro</h1>

      <img src="../assets/images/dashboard/incetro-logo.png" alt="">
    </div>

    <div class="dashboard-grid">
      <div @click="loadToSheets" class="card to-sheets">
        <Transition name="fade">
          <div v-if="toSheetsRequestProducer.state.value.loading" class="card-loader">
            <div class=" p-2 rounded-xl bg-blue-400 bg-opacity-25">
              <LoadingIcon class="text-4xl text-primary" />
            </div>
            <h4 class="text-center">Бот работает,<br>ожидайте</h4>
          </div>
        </Transition>
        <div class="card-header">
          <p>Из Notion в<br>Google Sheets</p>
          <LinkIcon class="text-2xl" />
        </div>

        <div class="footer text-white">
          <h3>Выгрузить</h3>
        </div>
      </div>

      <div class="card salary-notify p-default" @click="salaryNotify">
        <Transition name="fade">
          <div v-if="salaryNotifyRequestProducer.state.value.loading" class="card-loader">
            <div class=" p-2 rounded-xl bg-blue-400 bg-opacity-25">
              <LoadingIcon class="text-4xl text-primary" />
            </div>
            <h4 class="text-center">Бот работает, ожидайте</h4>
          </div>
        </Transition>

        <div class="card-header">
          <p>Провели<br>Выплаты ЗП?</p>
          <LinkIcon class="text-2xl" />
        </div>

        <div class="footer">
          <h3>Уведомить</h3>
        </div>
      </div>

      <div class="card qprogress">
        <div class="card-header">
          <p>Прогресс<br>(Q{{ currentQuarter }})</p>
          <LinkIcon class="text-2xl" />
        </div>

        <div class="footer">
          <h3>{{ quarterTasksDone }} / {{ querterTasks.length }}</h3>
          <div class="progress-bar mt-2">
            <div :style="{ width: `${(quarterTasksDone / querterTasks.length) * 100}%` }"></div>
          </div>
        </div>

      </div>

      <div class="card progress">
        <div class="card-header">
          <div class=" flex gap-1 items-center">
            <p>Мой прогресс</p>
          </div>
        </div>

        <div class="flex items-center gap-2 mt-2 mb-4">
          <h3>{{userTasks.filter(t => getCategoryByStatus(t.status) === TaskStatusCategory.Done).length}} / {{
            userTasks.length }}</h3>

          <p class=" flex items-center" :class="progressToStatus(userPercents)">
            <ArrowTopIcon class="arrow" /> {{ Math.round(Math.abs(userPercents)) }}%
          </p>
        </div>

        <div class="period-picker">
          <div @click="currentPeriod = Period.Quarter" class="period-picker-button"
            :class="{ active: currentPeriod == Period.Quarter }">
            Квартал
          </div>

          <div @click="currentPeriod = Period.Month" class="period-picker-button"
            :class="{ active: currentPeriod == Period.Month }">
            Месяц
          </div>

          <div @click="currentPeriod = Period.Week" class="period-picker-button"
            :class="{ active: currentPeriod == Period.Week }">
            Неделя
          </div>
        </div>
        <input type="file" ref="fileInput" @change="handleFileUpload" accept=".md" style="display: none;" />
      </div>

      <div class="row" id="row1">
        <div class="card mindmap-loader" @click="triggerFileUpload">
          <Transition name="fade">
            <div v-if="mindmapLoaderRequestProducer.state.value.loading" class="card-loader">
              <div class=" p-2 rounded-xl bg-blue-400 bg-opacity-25">
                <LoadingIcon class="text-4xl text-primary" />
              </div>
              <h4 class="text-center">Бот работает,<br>ожидайте</h4>
            </div>
          </Transition>

          <div class="card-header">
            <p>Майндкарты</p>
            <LinkIcon class="text-2xl" />
          </div>

          <div class="footer">
            <h3>Загрузить</h3>
          </div>
        </div>

        <div class="card feedback" @click="sendFeedback">
          <div class="card-header">
            <p>Есть идеи?</p>
            <LinkIcon class="text-2xl" />
          </div>

          <div class="footer">
            <h3>Фидбэк</h3>
          </div>
        </div>
      </div>

      <div class="row">
        <div class="card project-sheets-uploaders">
          <Transition name="fade">
            <div v-if="projectSheetsUpdateRequestProducer.state.value.loading" class="card-loader">
              <div class=" p-2 rounded-xl bg-blue-400 bg-opacity-25">
                <LoadingIcon class="text-4xl text-primary" />
              </div>
              <h4 class="text-center">Бот работает,<br>ожидайте</h4>
            </div>
          </Transition>

          <div class="card-header">
            <p>Выгрузка в Google Sheets</p>
            <!-- <LinkIcon class="text-2xl" /> -->
          </div>

          <div class="btns-list">
            <div class="project-el" v-for="project of projectsWithSheets" :key="project.id">
              <img class="project-icon" :src="project.icon" alt="" v-if="project.iconType !== 'emoji'">
              <p class="text-2xl" v-else>{{ project.icon }}</p>
              <a :href="project.sheetsLink" target="_blank">{{ project.name }}</a>
              <button class="px-4" @click="updateProjectSheets(project.id)">Синхронизация</button>
            </div>
          </div>
        </div>
      </div>

    </div>
  </main>
</template>

<style scoped>


.btns-list{
  @apply flex flex-col gap-2 w-full;
}
.project-el{
  @apply w-full flex gap-2 items-center;
}
.project-el a{
  @apply w-full underline overflow-hidden text-ellipsis whitespace-nowrap;
}
.project-icon{
  @apply w-8 h-8 rounded-full;
}

.messages {
  @apply fixed z-50 bottom-4 w-full h-fit flex flex-col items-center;
}

.message {
  @apply absolute bottom-0 bg-white rounded-2xl flex items-center gap-2 w-full p-4 shadow-lg;
  width: calc(100vw - 2rem);
}

.msg-icon {
  @apply p-2 rounded-xl flex justify-center items-center w-fit h-fit;
}

.message.error>.msg-icon {
  @apply text-danger bg-red-500 bg-opacity-25;
}

.message.success>.msg-icon {
  @apply text-success bg-green-500 bg-opacity-25;
}


.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.5s ease;
}

.fade-slide-enter-from,
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(10px);
}


.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.5s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

main {
  @apply p-small
}

.header {
  @apply flex justify-between items-center p-default;
}

.header>img {
  @apply h-10 w-10 rounded-full;
}

.dashboard-grid {
  @apply grid grid-cols-2 w-full gap-small;
}

.dashboard-grid .card {
  @apply rounded-2xl overflow-hidden relative bg-white p-default flex flex-col justify-between gap-8 min-h-40;
}

.card.project-sheets-uploaders{
  @apply h-auto min-h-0 gap-2;
}

.card.progress {
  @apply gap-0;
}

.row {
  @apply w-full;
  grid-column: 1/4;
}

.row#row1 {
  @apply flex gap-small;
}

.row#row1>div:first-child {
  @apply w-full;
}

.row#row1>div:last-child {
  @apply aspect-square h-40 w-40 min-w-40;
}


.dashboard-grid .card>.card-loader {
  @apply absolute w-full h-full bg-white bg-opacity-75 flex flex-col gap-4 justify-center items-center top-0 left-0;
  backdrop-filter: blur(0.5rem);
}

.card-header {
  @apply flex justify-between;
}

.to-sheets {
  grid-column: 1/2;
  grid-row: 1/3;
  background-image: url(../assets/images/dashboard/bg-waves.png);
  background-size: 125%;
  @apply bg-right-bottom bg-no-repeat bg-cover;
}

.salary-notify {
  grid-column: 2/3;
  grid-row: 1/2;
}

.qprogress {
  grid-column: 2/3;
  grid-row: 2/3;
}

.progress {
  @apply gap-0;
  grid-column: 1/3;
  grid-row: 3/4;
  min-height: auto !important;
}

.progress .good {
  @apply text-success;
}

.progress .bad>.arrow {
  @apply rotate-180;
}

.progress .neutral>.arrow {
  @apply hidden;
}

.progress .bad {
  @apply text-danger;
}

.progress .neutral {
  @apply text-warning;
}

.mindmap-loader {
  grid-column: 1/2;
  grid-row: 4/5;
}

.feedback {
  grid-column: 2/3;
  grid-row: 4/5;
  background-image: url(../assets/images/dashboard/bg-gradient.png);
  @apply bg-cover text-white;
}


.period-picker {
  @apply flex gap-1 p-0.5 bg-gray-200 rounded-full;
}

.period-picker>.period-picker-button {
  @apply w-full flex items-center justify-center p-1 rounded-full bg-transparent text-gray-400;
}

.period-picker>.period-picker-button.active {
  @apply bg-white text-black;
}

.progress-bar {
  @apply w-full h-8 bg-gray-300 rounded-full overflow-hidden;
}

.progress-bar>div {
  @apply bg-primary h-full;
}
</style>
